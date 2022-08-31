# demos

This directory contains scripts that demonstrate some possible uses of the
Registry API.

- [openapi.sh](openapi.sh) builds and analyzes a collection of APIs from the
  [OpenAPI Directory](https://github.com/APIs-guru/openapi-directory), a
  curated collection of OpenAPI descriptions of public APIs.
- [disco.sh](disco.sh) builds and analyzes a collection of APIs described by
  the [Google API Discovery Service](https://developers.google.com/discovery).
- [protos.sh](protos.sh) builds and analyzes a collection of APIs described by
  Protocol Buffer files distributed with the
  [googleapis](https://github.com/googleapis/googleapis) repo.

These scripts expect that the `PROJECT` is set to the id of the registry
project. For self-hosted (open source) installations, you can set this variable
by sourcing [setup-project.sh](setup-project.sh) as follows:

```
. setup-project.sh
```

This creates a project named `sample`.
