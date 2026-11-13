# docgen

Renders documentation from the compose files in `hosts/`, using Go templates.

The compose files are loaded with the official
[compose-go](https://github.com/compose-spec/compose-go) library, so `extends`,
`${VAR}` interpolation, port syntax and profiles behave as they do in Docker
Compose. Each stack gets its own `.env`. The values in `secrets.enc.env` are
never read: every key in it becomes a `${KEY}` placeholder, so a Traefik rule
like ``Host(`${WEBHOOK_HOST}.example.org`)`` shows up as written.

## Use

```sh
go -C tools/docgen run .          # render docs/templates/*.tmpl into docs/generated/
go -C tools/docgen run . -check   # exit 1 if docs/generated/ is out of date
go -C tools/docgen run . -json    # print the data templates receive
go -C tools/docgen test ./...
```

Every `docs/templates/NAME.tmpl` becomes `docs/generated/NAME`. Markdown output
gets a "generated, do not edit" comment on top. `docs/generated/` is in
`.prettierignore` because the tables are not column-aligned.

## Writing a template

Templates are standard `text/template`. Run with `-json` to see the data. The
top level has:

- `.Hosts`: host names.
- `.Services`: every service, sorted by host, stack and name. Each has `Host`,
  `Stack`, `Name`, `Container`, `Image`, `ImageRepo`, `ImageTag`, `ImageDigest`,
  `Restart`, `NetworkMode`, `Ports`, `Networks`, `Labels`, `Hostnames` and
  `File`.
- `.PortRows`: one row per published port, sorted by host and port, with
  `Host`, `Stack`, `Service`, `Container`, `Image`, `HostIP`, `Published`,
  `Target` and `Protocol`.

Extra functions: `join SEP LIST`, `code` (backticks, escapes `|`), `cell`
(escapes `|` and newlines) and `lower`.

Services with a compose profile are included. Ports without a host side
(`expose`) are not listed.
