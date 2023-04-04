# connection

This directory contains a Go package that can be used to get a Registry API
client that authenticates using passed Config.

Configurations can be automatically loaded from files in ~/.config/registry.
This includes the file `active_config` - which contains only the bare name (no
path) of a config file in the same directory. The config file it points to
should contain yaml configuration similar to this:

```yaml
registry:
  address: localhost:8080
  insecure: true
```

Note: These configuration files can be maintained by way of the
`registry config` and the `registry config configurations` commands. See the
help on those commands for more details.

The properties from these config files can be overridden using the following
flags:

```text
      --registry.address string   the server and port of the registry api (eg. localhost:8080)
      --registry.insecure         if specified, client connects via http (not https)
      --registry.token string     the token to use for authorization to registry
```

See `config.go` for more programming details.

The following environment variables are also used for overrides for testing and
internal purposes, but should be avoided for production as they are not
guaranteed:

- REGISTRY_ADDRESS
- REGISTRY_INSECURE
