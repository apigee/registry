# auth

This directory contains scripts that can be used to create registry
configurations containing credentials and server information for connections to
the `registry-server`.

- [LOCAL.sh](LOCAL.sh) creates a configuration ("local") that configures the
  registry client to work with a locally-running server.
- [HOSTED.sh](HOSTED.sh) creates a configuration ("hosted") that configures the
  registry client to work with a Google Cloud hosted server.

Each script should only be run once to create and activate a configuration.
Once created, the configurations can be selected be using the registry
commands. For example:

    $ registry config configurations activate local

Custom configurations can also be created:

    $ registry config configurations create custom --registry.address='myaddress' [etc...]

And any active configuration can be viewed and manipulated. Examples:

    $ registry config list
    $ registry config set registry.address="myaddress"
    $ registry config get registry.address

See for more details on creating and manipulating configurations:

    $ registry config --help
    $ registry config configurations --help
