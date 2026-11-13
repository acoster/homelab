package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Data is what the templates receive.
type Data struct {
	// Hosts are the directory names under hosts/, sorted.
	Hosts []string
	// Services lists every service of every stack, sorted by host, stack, name.
	Services []Service
	// PortRows has one row per published port, sorted by host, then port. It
	// saves a nested range in templates that want a "port per row" table.
	PortRows []PortRow
}

// Service is one compose service, fully resolved: extends merged, variables
// interpolated, profiles ignored (every service is included).
type Service struct {
	Host  string // directory under hosts/
	Stack string // directory under hosts/<host>/
	Name  string // service key in the compose file
	// Container is container_name, or the service name if none is set.
	Container   string
	Image       string // as written, including any digest
	ImageRepo   string // image without tag and digest
	ImageTag    string
	ImageDigest string
	Restart     string
	NetworkMode string
	Ports       []Port
	Networks    []Network
	Labels      map[string]string
	// Hostnames are the names in Traefik Host(`...`) rules of this service.
	Hostnames []string
	// File is the compose file, relative to the repo root.
	File string
}

// Port is a published port (one with a host side).
type Port struct {
	HostIP    string // empty when the port binds every address
	Published string
	Target    uint32
	Protocol  string
}

func (p Port) String() string {
	addr := p.Published
	if p.HostIP != "" {
		addr = p.HostIP + ":" + p.Published
	}
	return fmt.Sprintf("%s->%d/%s", addr, p.Target, p.Protocol)
}

// PortRow is a Port together with the service it belongs to.
type PortRow struct {
	Host, Stack, Service, Container, Image string
	Port
}

// Network is a service's attachment to a compose network.
type Network struct {
	Name string // key in the compose file's networks section
	IPv4 string
	IPv6 string
	MAC  string
}

func (d *Data) finish() {
	sort.Slice(d.Services, func(i, j int) bool {
		a, b := d.Services[i], d.Services[j]
		if a.Host != b.Host {
			return a.Host < b.Host
		}
		if a.Stack != b.Stack {
			return a.Stack < b.Stack
		}
		return a.Name < b.Name
	})
	for _, s := range d.Services {
		for _, p := range s.Ports {
			d.PortRows = append(d.PortRows, PortRow{s.Host, s.Stack, s.Name, s.Container, s.Image, p})
		}
	}
	sort.SliceStable(d.PortRows, func(i, j int) bool {
		a, b := d.PortRows[i], d.PortRows[j]
		if a.Host != b.Host {
			return a.Host < b.Host
		}
		if na, nb := portNumber(a.Published), portNumber(b.Published); na != nb {
			return na < nb
		}
		return a.Protocol < b.Protocol
	})
}

// portNumber is the first port of "8080" or "8000-8010"; 0 if unparsable.
func portNumber(s string) int {
	first, _, _ := strings.Cut(s, "-")
	n, _ := strconv.Atoi(first)
	return n
}

// splitImage breaks "repo:tag@sha256:..." into its parts. The registry may
// carry a port ("localhost:5000/app"), so a colon only starts the tag when
// it comes after the last slash.
func splitImage(image string) (repo, tag, digest string) {
	repo = image
	if i := strings.Index(repo, "@"); i >= 0 {
		repo, digest = repo[:i], repo[i+1:]
	}
	if i := strings.LastIndex(repo, ":"); i > strings.LastIndex(repo, "/") {
		repo, tag = repo[:i], repo[i+1:]
	}
	return repo, tag, digest
}
