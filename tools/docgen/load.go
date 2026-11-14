package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

var (
	// A key in a dotenv file. Values are ignored: in secrets.enc.env they are
	// ciphertext and must never reach the generated docs.
	dotenvKey   = regexp.MustCompile(`(?m)^([A-Za-z_][A-Za-z0-9_]*)=`)
	traefikHost = regexp.MustCompile("Host\\(`([^`]+)`\\)")
)

// Load reads every hosts/<host>/<stack>/compose.yaml below root.
func Load(ctx context.Context, root string) (*Data, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	files, err := filepath.Glob(filepath.Join(root, "hosts", "*", "*", "compose.yaml"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no hosts/*/*/compose.yaml below %s (is -root the repo root?)", root)
	}
	sort.Strings(files)

	data := &Data{}
	seen := map[string]bool{}
	for _, file := range files {
		stackDir := filepath.Dir(file)
		stack := filepath.Base(stackDir)
		host := filepath.Base(filepath.Dir(stackDir))
		rel, _ := filepath.Rel(root, file)
		rel = filepath.ToSlash(rel)

		services, err := loadStack(ctx, file)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rel, err)
		}
		for i := range services {
			services[i].Host, services[i].Stack, services[i].File = host, stack, rel
		}
		data.Services = append(data.Services, services...)
		if !seen[host] {
			seen[host] = true
			data.Hosts = append(data.Hosts, host)
		}
	}
	sort.Strings(data.Hosts)
	data.finish()
	return data, nil
}

// loadStack parses one compose file the way doco-cd would deploy it: with the
// stack's .env, and with every key of secrets.enc.env set to a "${KEY}"
// placeholder, because the real values are encrypted.
func loadStack(ctx context.Context, file string) ([]Service, error) {
	dir := filepath.Dir(file)
	opts := []cli.ProjectOptionsFn{
		cli.WithWorkingDirectory(dir),
		cli.WithName(filepath.Base(dir)),
		cli.WithEnv(secretPlaceholders(filepath.Join(dir, "secrets.enc.env"))),
	}
	if env := filepath.Join(dir, ".env"); exists(env) {
		opts = append(opts, cli.WithEnvFiles(env), cli.WithDotEnv)
	}
	opts = append(opts,
		cli.WithProfiles([]string{"*"}),  // document every service, whatever its profile
		cli.WithConsistency(false),       // external networks and the like need not resolve
		cli.WithoutEnvironmentResolution, // service env_files hold secrets; docs don't need them
	)

	po, err := cli.NewProjectOptions([]string{file}, opts...)
	if err != nil {
		return nil, err
	}
	project, err := cli.ProjectFromOptions(ctx, po)
	if err != nil {
		return nil, err
	}

	var out []Service
	for name, svc := range project.Services {
		out = append(out, convert(name, svc))
	}
	return out, nil
}

func convert(name string, svc types.ServiceConfig) Service {
	s := Service{
		Name:         name,
		Container:    svc.ContainerName,
		explicitName: svc.ContainerName,
		Image:        svc.Image,
		Restart:      svc.Restart,
		NetworkMode:  svc.NetworkMode,
		Labels:       map[string]string(svc.Labels),
	}
	if s.Container == "" {
		s.Container = name
	}
	s.ImageRepo, s.ImageTag, s.ImageDigest = splitImage(svc.Image)

	for _, p := range svc.Ports {
		if p.Published == "" { // "expose"-style entry, nothing on the host
			continue
		}
		s.Ports = append(s.Ports, Port{HostIP: p.HostIP, Published: p.Published, Target: p.Target, Protocol: p.Protocol})
	}

	for netName, n := range svc.Networks {
		net := Network{Name: netName}
		if n != nil {
			net.IPv4, net.IPv6, net.MAC = n.Ipv4Address, n.Ipv6Address, n.MacAddress
		}
		s.Networks = append(s.Networks, net)
	}
	sort.Slice(s.Networks, func(i, j int) bool { return s.Networks[i].Name < s.Networks[j].Name })

	hosts := map[string]bool{}
	for key, value := range svc.Labels {
		if !strings.HasPrefix(key, "traefik.http.routers.") || !strings.HasSuffix(key, ".rule") {
			continue
		}
		for _, m := range traefikHost.FindAllStringSubmatch(value, -1) {
			hosts[m[1]] = true
		}
	}
	for h := range hosts {
		s.Hostnames = append(s.Hostnames, h)
	}
	sort.Strings(s.Hostnames)
	return s
}

func secretPlaceholders(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var env []string
	seen := map[string]bool{}
	for _, m := range dotenvKey.FindAllStringSubmatch(string(content), -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			env = append(env, m[1]+"=${"+m[1]+"}")
		}
	}
	return env
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
