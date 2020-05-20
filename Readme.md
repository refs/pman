# ownCloud Infinite Scale: Runtime

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/8badecde63f743868c71850e43cdeb0d)](https://app.codacy.com/manual/refs_2/pman?utm_source=github.com&utm_medium=referral&utm_content=refs/pman&utm_campaign=Badge_Grade_Dashboard)

## Development

To run this project on binary mode:

```console
go install
pman // after this, the rpc service is ready to receive messages
```

on a different terminal session:

```console
pman --run phoenix
pman --run konnectd
pman --run proxy

pman --l

+-----------+-------+
| EXTENSION |  PID  |
+-----------+-------+
| konnectd  | 39950 |
| phoenix   | 39899 |
+-----------+-------+

pman --kill phoenix

+-----------+-------+
| EXTENSION |  PID  |
+-----------+-------+
| konnectd  | 39950 |
+-----------+-------+
```

## Security

If you find a security issue please contact [hello@zyxan.io](mailto:hello@zyxan.io) first.

## Contributing

Fork -> Patch -> Push -> Pull Request
