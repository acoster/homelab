package main

import (
	"fmt"
	"net/netip"
	"regexp"
	"sort"
	"strings"
)

// Lint finds conflicts that no single compose file shows, because each stack
// is validated on its own. It returns one sorted message per conflict:
//
//   - the same port published twice on one host (an unspecified bind address
//     overlaps every specific one)
//   - the same static IP, or MAC, on two services. IPs are LAN-wide, except
//     in Docker's default bridge range (172.16.0.0/12), where they only have
//     to be unique per host and network
//   - the same explicit container_name twice on one host
//   - the same Traefik router, service or middleware name defined by two
//     services. Traefik keeps one namespace per provider; this requires
//     names to be unique across every host, which is stricter but simple
//   - the same Traefik hostname served by two services. Several routers for
//     one host inside one service are normal (api / local variants).
//     Hostnames with a ${PLACEHOLDER} are skipped: their real value is
//     secret and differs per host.
//
// It cannot see anything outside the compose files: ports used by the OS or
// by network_mode: host containers, or addresses handed out by DHCP.
func Lint(d *Data) []string {
	var out []string
	out = append(out, lintPorts(d)...)
	out = append(out, lintAddresses(d)...)
	out = append(out, lintMACs(d)...)
	out = append(out, lintContainerNames(d)...)
	out = append(out, lintTraefikNames(d)...)
	out = append(out, lintHostnames(d)...)
	sort.Strings(out)
	return out
}

func (s Service) id() string { return s.Host + "/" + s.Stack + "/" + s.Name }

func lintPorts(d *Data) []string {
	type key struct{ host, port, proto string }
	type use struct{ bind, owner string }
	uses := map[key][]use{}
	for _, s := range d.Services {
		for _, p := range s.Ports {
			k := key{s.Host, p.Published, p.Protocol}
			uses[k] = append(uses[k], use{normalizeBind(p.HostIP), s.id()})
		}
	}

	var out []string
	for k, list := range uses {
		for i := range list {
			for j := i + 1; j < len(list); j++ {
				a, b := list[i], list[j]
				if a.bind != b.bind && a.bind != "*" && b.bind != "*" {
					continue
				}
				out = append(out, fmt.Sprintf("ports: %s publishes %s/%s twice: %s on %s and %s on %s",
					k.host, k.port, k.proto, a.owner, bindName(a.bind), b.owner, bindName(b.bind)))
			}
		}
	}
	return out
}

// normalizeBind maps every "all addresses" spelling to "*".
func normalizeBind(ip string) string {
	switch ip {
	case "", "0.0.0.0", "::", "[::]":
		return "*"
	}
	return ip
}

func bindName(bind string) string {
	if bind == "*" {
		return "all addresses"
	}
	return bind
}

// dockerBridges is where Docker allocates bridge networks by default.
var dockerBridges = netip.MustParsePrefix("172.16.0.0/12")

func lintAddresses(d *Data) []string {
	owners := map[string][]string{}
	label := map[string]string{}
	for _, s := range d.Services {
		for _, n := range s.Networks {
			for _, raw := range []string{n.IPv4, n.IPv6} {
				addr, err := netip.ParseAddr(raw)
				if err != nil {
					continue // empty, or a value we can't interpret
				}
				key, scope := addr.String(), "on the LAN"
				if dockerBridges.Contains(addr) {
					key = s.Host + "|" + n.Name + "|" + addr.String()
					scope = "on " + s.Host + " network " + n.Name
				}
				owners[key] = append(owners[key], s.id())
				label[key] = fmt.Sprintf("%s (%s)", addr, scope)
			}
		}
	}
	return duplicates("addresses", owners, func(key string, who []string) string {
		return fmt.Sprintf("%s is assigned to %s", label[key], strings.Join(who, " and "))
	})
}

func lintMACs(d *Data) []string {
	owners := map[string][]string{}
	for _, s := range d.Services {
		for _, n := range s.Networks {
			if n.MAC != "" {
				mac := strings.ToLower(n.MAC)
				owners[mac] = append(owners[mac], s.id())
			}
		}
	}
	return duplicates("macs", owners, func(mac string, who []string) string {
		return fmt.Sprintf("MAC %s is used by %s", mac, strings.Join(who, " and "))
	})
}

func lintContainerNames(d *Data) []string {
	owners := map[string][]string{}
	for _, s := range d.Services {
		if s.explicitName != "" {
			key := s.Host + "|" + s.explicitName
			owners[key] = append(owners[key], s.id())
		}
	}
	return duplicates("container_name", owners, func(key string, who []string) string {
		host, name, _ := strings.Cut(key, "|")
		return fmt.Sprintf("%s has container_name %q on %s", host, name, strings.Join(who, " and "))
	})
}

// traefikDefinition matches the labels that define a named Traefik object.
var traefikDefinition = regexp.MustCompile(`^traefik\.(http|tcp|udp)\.(routers|services|middlewares)\.([^.]+)\.`)

func lintTraefikNames(d *Data) []string {
	owners := map[string][]string{}
	for _, s := range d.Services {
		for label := range s.Labels {
			if m := traefikDefinition.FindStringSubmatch(label); m != nil {
				key := m[1] + " " + strings.TrimSuffix(m[2], "s") + " " + m[3]
				owners[key] = append(owners[key], s.id())
			}
		}
	}
	return duplicates("traefik", owners, func(key string, who []string) string {
		parts := strings.SplitN(key, " ", 3) // protocol, kind, name
		return fmt.Sprintf("%s %s %q is defined by %s", parts[0], parts[1], parts[2], strings.Join(who, " and "))
	})
}

func lintHostnames(d *Data) []string {
	owners := map[string][]string{}
	for _, s := range d.Services {
		for _, h := range s.Hostnames {
			if !strings.Contains(h, "${") {
				owners[h] = append(owners[h], s.id())
			}
		}
	}
	return duplicates("hostnames", owners, func(h string, who []string) string {
		return fmt.Sprintf("%s is served by %s", h, strings.Join(who, " and "))
	})
}

// duplicates reports every key with more than one distinct owner.
func duplicates(check string, owners map[string][]string, message func(key string, who []string) string) []string {
	var out []string
	for key, who := range owners {
		if who = unique(who); len(who) > 1 {
			out = append(out, check+": "+message(key, who))
		}
	}
	return out
}

func unique(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
