# disco-flame

This directory contains a modified version of the `disco` tool that is distributed with
[gnostic](https://github.com/googleapis/gnostic). It includes all the features of the
original tool plus support for an `--upload` flag that causes API descriptions to be
uploaded to a FLAME service.
  
Usage:

        disco-flame help

Prints a list of commands and options.

        disco-flame list [--raw]
        
Calls the Google Discovery API and lists available APIs. 
The `--raw` option prints the raw results of the Discovery List APIs call.

        disco-flame get [<api>] [<version>] [--upload] [--raw] [--openapi2] [--openapi3] [--features] [--schemas] [--all]
        
Gets the specified API and version from the Google Discovery API.
`<version>` can be omitted if it is unique.
The `--upload` option uploades the raw Discovery Format description to a FLAME service.
The `--raw` option saves the raw Discovery Format description.
The `--openapi2` option rewrites the API description in OpenAPI v2.
The `--openapi3` option rewrites the API description in OpenAPI v3.
The `--features` option displays the contents of the `features` sections of discovery documents.
The `--schemas` option displays information about the schemas defined for the API.
The `--all` option runs the other associated operations for all of the APIs available from the Discovery Service. 
When `--all` is specified, `<api>` and `<version>` should be omitted.

        disco-flame <file> [--upload] [--openapi2] [--openapi3] [--features] [--schemas]

Applies the specified operations to a local file. See the `get` command for details.

## The FLAME Service

Two environment variables are used to identify and to authenticate with the FLAME service.

- `CLI_FLAME_ADDRESS` is the address of the service, including a port number.
- `CLI_FLAME_TOKEN` is the authorization token of a service account that is able to call the API.

These environment variables are the same ones used to configure the FLAME CLI. For
more details, see the top-level README.