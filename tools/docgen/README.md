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
go -C tools/docgen run . -lint    # exit 1 on conflicting ports, addresses or names
go -C tools/docgen test ./...
```

To regenerate `docs/generated/` automatically, enable the repo's pre-commit
hook once per clone:

```sh
git config core.hooksPath .githooks
```

It runs the generator when a commit touches compose files, `.env` or secrets
key files, `common/`, templates or the tool itself, and adds the result to the
same commit. Skip it for one commit with `git commit --no-verify`; CI still
checks the docs.

Every `docs/templates/NAME.tmpl` becomes `docs/generated/NAME`. Markdown output
gets a "generated, do not edit" comment on top. `docs/generated/` is in
`.prettierignore` because the tables are not column-aligned.

## Lint

`-lint` checks things no single compose file can show, because each stack is
validated on its own:

- The same port published twice on one host. A bind to all addresses overlaps
  every specific address.
- The same static IP or MAC on two services. IPs are LAN-wide, except in
  Docker's default bridge range (172.16.0.0/12), where they only need to be
  unique per host and network.
- The same `container_name` twice on one host.
- The same Traefik router, service or middleware name defined by two services.
  This is stricter than Traefik, which keeps a namespace per provider, but it
  keeps names safe to move between hosts.
- The same Traefik hostname served by two services. Several routers for one
  host inside one service are fine. Hostnames with a `${PLACEHOLDER}` are
  skipped, because the real value is secret and differs per host.

It cannot see anything outside the compose files: ports used by the OS or by
`network_mode: host` containers, or addresses handed out by DHCP.

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
