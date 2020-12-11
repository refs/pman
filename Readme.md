# ownCloud Infinite Scale: Runtime

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/8badecde63f743868c71850e43cdeb0d)](https://app.codacy.com/manual/refs_2/pman?utm_source=github.com&utm_medium=referral&utm_content=refs/pman&utm_campaign=Badge_Grade_Dashboard)
[![Go Reference](https://pkg.go.dev/badge/github.com/refs/pman.svg)](https://pkg.go.dev/github.com/refs/pman)
[![Release](https://img.shields.io/github/release/refs/pman.svg?style=flat-square)](https://github.com/refs/pman/releases/latest)

## Development

To run this project on binary mode:

```console
go install
pman // after this, the rpc service is ready to receive messages
```

on a different terminal session:

```console
pman run phoenix
pman run konnectd
pman run proxy

pman list

+--------------------------+-------+
|        EXTENSION         |  PID  |
+--------------------------+-------+
| konnectd                 | 67556 |
| phoenix                  | 67537 |
| proxy                    | 67535 |
+--------------------------+-------+

```

## Security

If you find a security issue please contact [hello@zyxan.io](mailto:hello@zyxan.io) first.

## Contributing

Fork -> Patch -> Push -> Pull Request
