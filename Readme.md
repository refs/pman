# ownCloud Infinite Scale: Runtime

## Development

To run this project on binary mode":

```console
go install
pman // after this, the rpc service is ready to receive messages
```

on a different terminal session:

```console
pman --run phoenix
pman --run konnectd
pman --run proxy
```

## Security

If you find a security issue please contact hello@zyxan.io first.

## Contributing

Fork -> Patch -> Push -> Pull Request
