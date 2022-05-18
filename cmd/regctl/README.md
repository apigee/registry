# regctl

`regctl` is a command line interface for the Registry API that provides a
mixture of high-level handwritten features and automatically-generated
subcommands for calling the rpc services in the Registry API.

## Usage

Run `regctl help` for general information and `regctl help [subcommand]` for
details on any subcommand.

## Configuration

`regctl` is configured with environment variables that are set by authorization
scripts in the [auth](/auth) directory of this project. This includes the
address of the Registry API server and authentication tokens.

## Release builds

Release builds of the `regctl` tool are available on GitHub and can be found in
[github.com/apigee/registry/releases](https://github.com/apigee/registry/releases).

**MacOS note:** To run the `regctl` tool on MacOS, you may need to
[unquarantine](https://discussions.apple.com/thread/3145071) it by running the
following on the command line:

```
xattr -d com.apple.quarantine regctl
```

## Running with Apigee API hub

To use the `regctl` tool with a hosted instance associated with Apigee API hub,
please do the following:

1. Make sure you have gcloud command installed.
2. Set the `PROJECT_ID` environment variable to your API hub project name.
3. Configure `gcloud` to use your project:

```
gcloud config set project $PROJECT_ID
```

4. Run the following script to get an authorization token and set it in your
   environment as `$APG_REGISTRY_TOKEN` (on Windows, please use
   `auth/HOSTED.bat`):

```
. auth/HOSTED.sh
```

5. To list all the APIs in your API hub instance run the following:

```
regctl list projects/$PROJECT_ID/locations/global/apis/-
```

6. To see other supported commands, run the following:

```
regctl help
```

## Importing API information with the Registry tool

The `regctl` tool currently includes several demonstration subcommands that
upload API descriptions into a registry and one subcommmand (`regctl apply`)
that is likely to become a recommended way to populate an API registry.

- `regctl upload bulk openapi` reads OpenAPI descriptions from a directory that
  follows the style of the
  [APIs-guru/openapi-directory](https://github.com/APIs-guru/openapi-directory)
  repository. To try it, clone the `openapi-directory` repo, change your
  directory to the repo, and run the `regctl upload bulk openapi` command as
  follows:

  ```
  git clone https://github.com/apis-guru/openapi-directory
  cd openapi-directory
  regctl upload bulk openapi APIs --project-id $PROJECT_ID
  ```

  Here `APIs` is a directory in the repo and `$PROJECT_ID` should be set to
  your registry project id.

- `regctl upload bulk protos` reads Protocol Buffer API descriptions from a
  directory that follows the style of the
  [googleapis/googleapis](https://github.com/googleapis/googleapis) repository.
  To try it, clone the `googleapis` repo, change your directory to the root of
  the repo, and run the `regctl upload bulk protos` command as follows:

  ```
  git clone https://github.com/googleapis/googleapis
  cd googleapis
  regctl upload bulk protos . --project-id $PROJECT_ID
  ```

  As above, `$PROJECT_ID` should be set to your registry project id.

- `regctl upload bulk discovery` reads API descriptions from the
  [Google API Discovery Service](https://developers.google.com/discovery). This
  reads from an online service, so you can try it by simply running the
  following:

  ```
  regctl upload bulk discovery --project-id $PROJECT_ID
  ```

  As above, `$PROJECT_ID` should be set to your registry project id.

- `regctl apply` reads API information from YAML files using a mechanism
  similar to `kubectl apply`. For details, see
  [this GitHub issue](https://github.com/apigee/registry/issues/450). To try
  it, run the following from the root of the `registry` repo:

  ```
  regctl apply -f cmd/regctl/cmd/apply/testdata/registry.yaml --parent projects/$PROJECT_ID/locations/global
  ```
