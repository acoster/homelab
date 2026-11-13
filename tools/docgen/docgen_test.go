package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fixture builds a tiny repo: one shared base service, one host with two
// stacks, a .env, and a secrets file whose values must never be shown.
func fixture(t *testing.T) string {
	root := t.TempDir()
	write(t, root, "common/compose.base.yaml", `
services:
  base:
    image: example/base:1.2@sha256:abc
    restart: unless-stopped
`)
	write(t, root, "hosts/h1/web/.env", "HOST_IP=10.0.0.1\nDOMAIN=example.org\n")
	write(t, root, "hosts/h1/web/secrets.enc.env", "SECRET_HOST=ENC[AES256_GCM,data:ciphertext]\nTOKEN=ENC[AES256_GCM,data:other]\n")
	write(t, root, "hosts/h1/web/compose.yaml", `
name: web
services:
  app:
    extends:
      file: ../../../common/compose.base.yaml
      service: base
    container_name: web-app
    ports:
      - '${HOST_IP:?}:8080:80'
    labels:
      traefik.http.routers.app.rule: Host(`+"`app.${DOMAIN}`"+`) || Host(`+"`${SECRET_HOST}.${DOMAIN}`"+`)
    networks:
      lan:
        ipv4_address: 192.168.1.50
  debug:
    image: localhost:5000/tool
    profiles: [debug]
    ports:
      - '9090:90/udp'
    environment:
      TOKEN: ${TOKEN}
networks:
  lan:
    external: true
`)
	write(t, root, "hosts/h1/db/compose.yaml", `
services:
  postgres:
    image: postgres:17
    expose: ['5432']
`)
	return root
}

func TestLoad(t *testing.T) {
	data, err := Load(context.Background(), fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(data.Hosts, ","); got != "h1" {
		t.Fatalf("hosts = %q", got)
	}

	var names []string
	for _, s := range data.Services {
		names = append(names, s.Stack+"/"+s.Name)
	}
	if got, want := strings.Join(names, " "), "db/postgres web/app web/debug"; got != want {
		t.Fatalf("services = %q, want %q (sorted, profiled service included)", got, want)
	}

	app := data.Services[1]
	if app.Image != "example/base:1.2@sha256:abc" || app.ImageRepo != "example/base" || app.ImageTag != "1.2" || app.ImageDigest != "sha256:abc" {
		t.Errorf("image from extends not split correctly: %+v", app)
	}
	if app.Restart != "unless-stopped" || app.Container != "web-app" {
		t.Errorf("restart/container = %q/%q", app.Restart, app.Container)
	}
	if len(app.Ports) != 1 || app.Ports[0].String() != "10.0.0.1:8080->80/tcp" {
		t.Errorf("ports = %v", app.Ports)
	}
	if got := strings.Join(app.Hostnames, ","); got != "${SECRET_HOST}.example.org,app.example.org" {
		t.Errorf("hostnames = %q (secret values must stay placeholders)", got)
	}
	if len(app.Networks) != 1 || app.Networks[0].IPv4 != "192.168.1.50" {
		t.Errorf("networks = %+v", app.Networks)
	}
	if app.File != "hosts/h1/web/compose.yaml" {
		t.Errorf("file = %q", app.File)
	}

	if pg := data.Services[0]; pg.Container != "postgres" || len(pg.Ports) != 0 {
		t.Errorf("expose-only service should have no published ports: %+v", pg)
	}
	if dbg := data.Services[2]; dbg.ImageRepo != "localhost:5000/tool" || dbg.ImageTag != "" {
		t.Errorf("registry with a port mis-split: %+v", dbg)
	}

	if len(data.PortRows) != 2 || data.PortRows[0].Published != "8080" || data.PortRows[1].Protocol != "udp" {
		t.Errorf("port rows = %+v", data.PortRows)
	}
}

func TestRender(t *testing.T) {
	root := fixture(t)
	write(t, root, "docs/templates/ports.md.tmpl",
		"{{ range .PortRows }}{{ .Host }} {{ code .HostIP }} {{ .Published }} {{ .Container }}\n{{ end }}")
	data, err := Load(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	out, err := RenderAll(data, filepath.Join(root, "docs/templates"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(out["ports.md"])
	if !strings.HasPrefix(got, "<!-- Generated") {
		t.Errorf("missing generated-file header:\n%s", got)
	}
	for _, want := range []string{"h1 `10.0.0.1` 8080 web-app\n", "h1  9090 debug\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("output lacks %q:\n%s", want, got)
		}
	}
}

func TestRunCheck(t *testing.T) {
	root := fixture(t)
	write(t, root, "docs/templates/x.txt.tmpl", "{{ len .Services }}\n")
	if err := run(root, "docs/templates", "docs/generated", true, false); err == nil {
		t.Fatal("-check should fail before anything is generated")
	}
	if err := run(root, "docs/templates", "docs/generated", false, false); err != nil {
		t.Fatal(err)
	}
	if err := run(root, "docs/templates", "docs/generated", true, false); err != nil {
		t.Fatalf("-check should pass right after generating: %v", err)
	}
}

func TestFindRoot(t *testing.T) {
	root := fixture(t)
	sub := filepath.Join(root, "hosts", "h1", "web")
	t.Chdir(sub)
	got, err := findRoot()
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.EvalSymlinks(root)
	if got, _ = filepath.EvalSymlinks(got); got != want {
		t.Errorf("findRoot() = %q, want %q", got, want)
	}
}
