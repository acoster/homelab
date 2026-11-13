// Command docgen renders documentation from the compose files in this repo.
//
//	go -C tools/docgen run .          write docs/generated/* from docs/templates/*.tmpl
//	go -C tools/docgen run . -check   fail if docs/generated/* is out of date
//	go -C tools/docgen run . -json    print the data the templates receive
//
// The repo root is found by walking up from the current directory to the
// first one that has a hosts/ directory; -root overrides that.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	root := flag.String("root", "", "repo root (default: the nearest parent directory with a hosts/ directory)")
	templates := flag.String("templates", "docs/templates", "directory with *.tmpl files, relative to -root")
	outDir := flag.String("out", "docs/generated", "output directory, relative to -root")
	check := flag.Bool("check", false, "write nothing; exit 1 if the output files differ from what would be generated")
	dump := flag.Bool("json", false, "print the template data as JSON and exit")
	flag.Parse()

	if err := run(*root, *templates, *outDir, *check, *dump); err != nil {
		fmt.Fprintln(os.Stderr, "docgen:", err)
		os.Exit(1)
	}
}

func run(root, templates, outDir string, check, dump bool) error {
	if root == "" {
		var err error
		if root, err = findRoot(); err != nil {
			return err
		}
	}
	data, err := Load(context.Background(), root)
	if err != nil {
		return err
	}
	if dump {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	rendered, err := RenderAll(data, filepath.Join(root, templates))
	if err != nil {
		return err
	}

	names := make([]string, 0, len(rendered))
	for name := range rendered {
		names = append(names, name)
	}
	sort.Strings(names)

	stale := false
	for _, name := range names {
		path := filepath.Join(root, outDir, name)
		current, _ := os.ReadFile(path)
		if bytes.Equal(current, rendered[name]) {
			continue
		}
		if check {
			fmt.Fprintf(os.Stderr, "%s is out of date\n", filepath.Join(outDir, name))
			stale = true
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, rendered[name], 0o644); err != nil {
			return err
		}
		fmt.Println("wrote", filepath.Join(outDir, name))
	}
	if stale {
		return fmt.Errorf("run `go -C tools/docgen run .` and commit the result")
	}
	return nil
}

// findRoot walks up from the working directory to the repo root, which is the
// first directory containing hosts/.
func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "hosts")); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no hosts/ directory in the working directory or any parent; run inside the repo or pass -root")
		}
		dir = parent
	}
}
