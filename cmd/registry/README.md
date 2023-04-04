# registry

`registry` is a command-line tool for working with the Registry API that
provides a mixture of high-level capabilities and automatically-generated
subcommands that call methods of the Registry API.

## Usage

Run `registry help` for general information and `registry help [subcommand]` for
details on any subcommand.

## Configuration

`registry` is configured with environment variables that are set by
authorization scripts in the [auth](/auth) directory of this project. This
includes the address of the Registry API server and authentication tokens.

## Release builds

Release builds of the `registry` tool are available on GitHub and can be found
in
[github.com/apigee/registry/releases](https://github.com/apigee/registry/releases).

**MacOS note:** To run the `registry` tool on MacOS, you may need to
[unquarantine](https://discussions.apple.com/thread/3145071) it by running the
following on the command line:

```
xattr -d com.apple.quarantine registry
```

## Easy installation

This will download and install the latest version of the `registry` tool (on
Linux or Darwin):

```
curl -L https://raw.githubusercontent.com/apigee/registry/main/downloadLatest.sh | sh -
```

## Running with Apigee API hub

To use the `registry` tool with a hosted instance associated with Apigee API
hub, please do the following:

1. Make sure you have gcloud command installed.
2. Set the `PROJECT_ID` environment variable to your API hub project name.
3. Configure `gcloud` to use your project:

```
gcloud config set project $PROJECT_ID
```

4. Run the following script to get an authorization token and set it in your
   environment as `$REGISTRY_TOKEN` (on Windows, please use `auth/HOSTED.bat`):

```
. auth/HOSTED.sh
```

5. To list all the APIs in your API hub instance run the following:

```
registry get projects/$PROJECT_ID/locations/global/apis/-
```

6. To see other supported commands, run the following:

```
registry help
```

## Importing API information with the Registry tool

The `registry` tool currently includes several demonstration subcommands that
upload API descriptions into a registry and one subcommmand (`registry apply`)
that is likely to become a recommended way to populate an API registry.

- `registry upload openapi` reads OpenAPI descriptions from a directory that
  follows the style of the
  [APIs-guru/openapi-directory](https://github.com/APIs-guru/openapi-directory)
  repository. To try it, clone the `openapi-directory` repo, change your
  directory to the repo, and run the `registry upload openapi` command as
  follows:

  ```
  git clone https://github.com/apis-guru/openapi-directory
  cd openapi-directory
  registry upload openapi APIs --project-id $PROJECT_ID
  ```

  Here `APIs` is a directory in the repo and `$PROJECT_ID` should be set to your
  registry project id.

- `registry upload protos` reads Protocol Buffer API descriptions from a
  directory that follows the style of the
  [googleapis/googleapis](https://github.com/googleapis/googleapis) repository.
  To try it, clone the `googleapis` repo, change your directory to the root of
  the repo, and run the `registry upload protos` command as follows:

  ```
  git clone https://github.com/googleapis/googleapis
  cd googleapis
  registry upload protos . --project-id $PROJECT_ID
  ```

  As above, `$PROJECT_ID` should be set to your registry project id.

- `registry upload discovery` reads API descriptions from the
  [Google API Discovery Service](https://developers.google.com/discovery). This
  reads from an online service, so you can try it by simply running the
  following:

  ```
  registry upload discovery --project-id $PROJECT_ID
  ```

  As above, `$PROJECT_ID` should be set to your registry project id.

- `registry apply` reads API information from YAML files using a mechanism
  similar to `kubectl apply`. For details,
  [check the wiki entry](https://github.com/apigee/registry/wiki/registry-apply)
  To try it, run the following from the root of the `registry` repo:

  ```
  registry apply -f cmd/registry/cmd/apply/testdata/registry.yaml --parent projects/$PROJECT_ID/locations/global
  ```
