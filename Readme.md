# ownCloud Infinite Scale: Runtime

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/8badecde63f743868c71850e43cdeb0d)](https://app.codacy.com/manual/refs_2/pman?utm_source=github.com&utm_medium=referral&utm_content=refs/pman&utm_campaign=Badge_Grade_Dashboard)
[![Go Reference](https://pkg.go.dev/badge/github.com/refs/pman.svg)](https://pkg.go.dev/github.com/refs/pman)
[![Release](https://img.shields.io/github/release/refs/pman.svg?style=flat-square)](https://github.com/refs/pman/releases/latest)

Pman is a slim utility library for supervising long-running processes. It can be [embedded](https://github.com/owncloud/OCIS/blob/ea2a2b328e7261ed72e65adf48359c0a44e14b40/OCIS/pkg/runtime/runtime.go#L84) or used as a cli command.

When used as a CLI command it relays actions to a running runtime.

## Usage

Start a runtime

```go
package main
import "github.com/refs/pman/pkg/service"

func main() {
    service.Start()    
}
```
![start runtime](https://imgur.com/F67hgQk.gif)

Start sending messages
![message runtime](https://imgur.com/O71RlsJ.gif)

## Security

If you find a security issue please contact [hello@zyxan.io](mailto:hello@zyxan.io) ffirst.

## Contributing

Fork -> Patch -> Push -> Pull Request
