package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// Functions available in templates, next to the standard ones (range, with,
// printf, len, index, ...).
var funcs = template.FuncMap{
	// join SEP LIST, so it also works in a pipeline: {{ .Hostnames | join ", " }}
	"join": func(sep string, elems []string) string { return strings.Join(elems, sep) },
	// code wraps in backticks (empty stays empty) and escapes | for tables.
	"code": func(s string) string {
		if s == "" {
			return ""
		}
		return "`" + strings.ReplaceAll(s, "|", `\|`) + "`"
	},
	// cell makes plain text safe inside a table cell.
	"cell": func(s string) string {
		return strings.NewReplacer("|", `\|`, "\n", " ").Replace(s)
	},
	"lower": strings.ToLower,
}

// RenderAll renders every *.tmpl in dir. The result maps the output file name
// (the template name without ".tmpl") to its content.
func RenderAll(data *Data, dir string) (map[string][]byte, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.tmpl"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no *.tmpl files in %s", dir)
	}
	out := map[string][]byte{}
	for _, file := range files {
		base := filepath.Base(file)
		tmpl, err := template.New(base).Funcs(funcs).Option("missingkey=error").ParseFiles(file)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		name := strings.TrimSuffix(base, ".tmpl")
		if strings.HasSuffix(name, ".md") {
			fmt.Fprintf(&buf, "<!-- Generated from %s by tools/docgen. Do not edit. -->\n\n", filepath.ToSlash(filepath.Join(dir, base)))
		}
		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("%s: %w", base, err)
		}
		out[name] = buf.Bytes()
	}
	return out, nil
}
