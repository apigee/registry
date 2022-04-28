# registry

`registry` is a handwritten command line interface for the Registry API that
provides high-level features that are not in the automatically-generated `apg`
tool.

## Usage

Run `registry help` for general information and `registry help [subcommand]`
for details on any subcommand.

## Configuration

`registry` is configured with environment variables that are set by
authorization scripts in the [auth](/auth) directory of this project. This
includes the address of the Registry API server and authentication tokens.

## Release builds

Release builds of the `registry` tool are available on GitHub and can be found
in
[github.com/apigee/registry/releases](https://github.com/apigee/registry/releases).

**MacOS note:** To run the `registry` tool on MacOS, you will need to
unquarantine it by running the following on the command line:

> xattr -d com.apple.quarantine registry

## Running with Apigee API hub

To use the `registry` tool with a hosted instance associated with Apigee API
hub, please do the following:

1. Make sure you have gcloud command installed.
2. Set the GCP_PROJECT environment variable to your API hub project name.
3. Configure `gcloud` to use your project
   > gcloud config set project \$GCP_PROJECT
4. Run the following script to get an authorization token and set it in your
   environment.
   > . auth/HOSTED.sh (On Windows, please use `auth/HOSTED.bat`.)
5. To list all the APIs in your API hub instance run the following:
   > registry list projects/\$GCP_PROJECT/locations/global/apis/-
6. To see other supported commands, run the following:
   > registry help
