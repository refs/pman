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
pman run phoenix
pman run konnectd
pman run proxy

pman list

+--------------------------+-------+
|        EXTENSION         |  PID  |
+--------------------------+-------+
| accounts                 | 67554 |
| api                      | 67558 |
| glauth                   | 67555 |
| graph                    | 67538 |
| graph-explorer           | 67539 |
| konnectd                 | 67556 |
| ocs                      | 67540 |
| phoenix                  | 67537 |
| proxy                    | 67535 |
| registry                 | 67560 |
| reva-auth-basic          | 67545 |
| reva-auth-bearer         | 67546 |
| reva-frontend            | 67542 |
| reva-gateway             | 67543 |
| reva-sharing             | 67561 |
| reva-storage-eos         | 67550 |
| reva-storage-eos-data    | 67551 |
| reva-storage-home        | 67547 |
| reva-storage-home-data   | 67549 |
| reva-storage-oc          | 67552 |
| reva-storage-oc-data     | 67553 |
| reva-storage-public-link | 67548 |
| reva-users               | 67544 |
| settings                 | 67536 |
| thumbnails               | 67557 |
| web                      | 67559 |
| webdav                   | 67541 |
+--------------------------+-------+

```

## Security

If you find a security issue please contact [hello@zyxan.io](mailto:hello@zyxan.io) first.

## Contributing

Fork -> Patch -> Push -> Pull Request
