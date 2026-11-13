package main

import (
	"context"
	"strings"
	"testing"
)

// svc builds a service; edit adjusts it.
func svc(host, stack, name string, edit func(*Service)) Service {
	s := Service{Host: host, Stack: stack, Name: name, Container: name}
	if edit != nil {
		edit(&s)
	}
	return s
}

func port(ip, published, proto string) Port {
	return Port{HostIP: ip, Published: published, Target: 80, Protocol: proto}
}

func TestLint(t *testing.T) {
	withPorts := func(ports ...Port) func(*Service) { return func(s *Service) { s.Ports = ports } }
	withNet := func(name, ipv4, mac string) func(*Service) {
		return func(s *Service) { s.Networks = []Network{{Name: name, IPv4: ipv4, MAC: mac}} }
	}
	withLabels := func(kv ...string) func(*Service) {
		return func(s *Service) {
			s.Labels = map[string]string{}
			for i := 0; i < len(kv); i += 2 {
				s.Labels[kv[i]] = kv[i+1]
			}
		}
	}
	withHosts := func(h ...string) func(*Service) { return func(s *Service) { s.Hostnames = h } }
	named := func(n string) func(*Service) { return func(s *Service) { s.explicitName = n } }

	tests := []struct {
		name     string
		services []Service
		want     []string // substrings, one per expected finding; empty means clean
	}{
		{"clean", []Service{svc("h", "a", "x", withPorts(port("10.0.0.1", "80", "tcp"))), svc("h", "b", "y", withPorts(port("10.0.0.1", "81", "tcp")))}, nil},

		{"same ip and port on one host",
			[]Service{svc("h", "a", "x", withPorts(port("10.0.0.1", "80", "tcp"))), svc("h", "b", "y", withPorts(port("10.0.0.1", "80", "tcp")))},
			[]string{"ports: h publishes 80/tcp twice: h/a/x on 10.0.0.1 and h/b/y on 10.0.0.1"}},
		{"unspecified bind overlaps a specific ip",
			[]Service{svc("h", "a", "x", withPorts(port("", "80", "tcp"))), svc("h", "b", "y", withPorts(port("10.0.0.1", "80", "tcp")))},
			[]string{"on all addresses"}},
		{"0.0.0.0 counts as unspecified",
			[]Service{svc("h", "a", "x", withPorts(port("0.0.0.0", "80", "tcp"))), svc("h", "b", "y", withPorts(port("", "80", "tcp")))},
			[]string{"ports:"}},
		{"different ips, protocols or hosts are fine",
			[]Service{
				svc("h", "a", "x", withPorts(port("10.0.0.1", "80", "tcp"), port("10.0.0.1", "80", "udp"))),
				svc("h", "b", "y", withPorts(port("10.0.0.2", "80", "tcp"))),
				svc("g", "a", "z", withPorts(port("10.0.0.1", "80", "tcp"))),
			}, nil},

		{"lan ip reused across hosts",
			[]Service{svc("h", "a", "x", withNet("lan", "192.168.1.5", "")), svc("g", "a", "y", withNet("lan", "192.168.1.5", ""))},
			[]string{"192.168.1.5 (on the LAN) is assigned to g/a/y and h/a/x"}},
		{"bridge ip reused on another host is fine",
			[]Service{svc("h", "a", "x", withNet("servers", "172.18.0.5", "")), svc("g", "a", "y", withNet("servers", "172.18.0.5", ""))}, nil},
		{"bridge ip reused on the same host network",
			[]Service{svc("h", "a", "x", withNet("servers", "172.18.0.5", "")), svc("h", "b", "y", withNet("servers", "172.18.0.5", ""))},
			[]string{"172.18.0.5 (on h network servers)"}},
		{"one service may keep its own ip",
			[]Service{svc("h", "a", "x", func(s *Service) {
				s.Networks = []Network{{Name: "a", IPv4: "192.168.1.5"}, {Name: "b", IPv4: "192.168.1.5"}}
			})}, nil},

		{"mac matches regardless of case",
			[]Service{svc("h", "a", "x", withNet("lan", "", "4C:81:BE:00:00:01")), svc("g", "a", "y", withNet("lan", "", "4c:81:be:00:00:01"))},
			[]string{"macs: MAC 4c:81:be:00:00:01 is used by g/a/y and h/a/x"}},
		{"one service may reuse its own mac",
			[]Service{svc("h", "a", "x", func(s *Service) {
				s.Networks = []Network{{Name: "a", MAC: "aa:aa:aa:aa:aa:aa"}, {Name: "b", MAC: "aa:aa:aa:aa:aa:aa"}}
			})}, nil},

		{"container_name twice on a host",
			[]Service{svc("h", "a", "x", named("db")), svc("h", "b", "y", named("db"))},
			[]string{`h has container_name "db" on h/a/x and h/b/y`}},
		{"container_name on different hosts, or defaulted, is fine",
			[]Service{svc("h", "a", "x", named("db")), svc("g", "a", "y", named("db")), svc("h", "b", "db", nil), svc("h", "c", "db", nil)}, nil},

		{"router defined by two services",
			[]Service{svc("h", "a", "x", withLabels("traefik.http.routers.app.rule", "r")), svc("g", "a", "y", withLabels("traefik.http.routers.app.rule", "r"))},
			[]string{`traefik: http router "app" is defined by g/a/y and h/a/x`}},
		{"middleware and service names are checked too",
			[]Service{
				svc("h", "a", "x", withLabels("traefik.http.middlewares.auth.headers.x", "1", "traefik.http.services.s.loadbalancer.server.port", "1")),
				svc("g", "a", "y", withLabels("traefik.http.middlewares.auth.headers.x", "1", "traefik.http.services.s.loadbalancer.server.port", "1")),
			},
			[]string{"http middleware", "http service"}},
		{"many routers in one service, and router vs service of the same name, are fine",
			[]Service{svc("h", "a", "x", withLabels(
				"traefik.http.routers.app.rule", "r", "traefik.http.routers.app-api.rule", "r", "traefik.http.services.app.loadbalancer.server.port", "1"))}, nil},

		{"hostname served by two services",
			[]Service{svc("h", "a", "x", withHosts("a.example.org")), svc("g", "a", "y", withHosts("a.example.org"))},
			[]string{"hostnames: a.example.org is served by g/a/y and h/a/x"}},
		{"hostname repeated inside one service, or with a placeholder, is fine",
			[]Service{svc("h", "a", "x", withHosts("a.example.org", "a.example.org")), svc("h", "b", "y", withHosts("${WEBHOOK_HOST}.example.org")), svc("g", "b", "z", withHosts("${WEBHOOK_HOST}.example.org"))}, nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Lint(&Data{Services: tc.services})
			if len(got) != len(tc.want) {
				t.Fatalf("got %d finding(s), want %d:\n%s", len(got), len(tc.want), strings.Join(got, "\n"))
			}
			for _, want := range tc.want {
				found := false
				for _, g := range got {
					found = found || strings.Contains(g, want)
				}
				if !found {
					t.Errorf("no finding contains %q; got:\n%s", want, strings.Join(got, "\n"))
				}
			}
		})
	}
}

func TestLintFixtureIsClean(t *testing.T) {
	data, err := Load(context.Background(), fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if got := Lint(data); len(got) != 0 {
		t.Fatalf("unexpected findings:\n%s", strings.Join(got, "\n"))
	}
}

func TestRunLint(t *testing.T) {
	root := fixture(t)
	if err := run(root, "docs/templates", "docs/generated", false, false, true); err != nil {
		t.Fatalf("clean fixture should lint clean: %v", err)
	}
	// A second stack publishing the same address and port as web/app.
	write(t, root, "hosts/h1/other/compose.yaml", "services:\n  clash:\n    image: x\n    ports: ['10.0.0.1:8080:80']\n")
	write(t, root, "hosts/h1/other/.env", "")
	if err := run(root, "docs/templates", "docs/generated", false, false, true); err == nil {
		t.Fatal("conflicting ports should fail -lint")
	}
}
