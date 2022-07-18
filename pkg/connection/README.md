# connection

This directory contains a Go package that can be used to get a Registry API
client that authenticates using passed Settings.

Settings can be automatically loaded from files in ~/.config/registry. This
includes the file `active_config` - which contains only the bare name (no path)
of a config file in the same directory. The config file it points to should
contain yaml configuration similar to this:

``` yaml
registry:
    address: localhost:8080
    insecure: true
```

See `settings.go` for more details.

The following environment variables can also be used instead of a config file
(or as overrides) for internal purposes, but should be avoided for production
purposes as they may be removed at any time:

- APG_REGISTRY_ADDRESS
- APG_REGISTRY_INSECURE
- APG_REGISTRY_TOKEN
