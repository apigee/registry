#!/bin/bash
#
# Copyright 2021 Google LLC. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# This script uses the [googleapis](https://github.com/googleapis/googleapis)
# Protocol Buffer descriptions of public Google APIs to build a collection
# of API descriptions and them performs some analysis on them.
# It assumes that the repo has been cloned to the user's Desktop,
# so that it can be found at `~/Desktop/googleapis`.

# It also assumes an enviroment configured to call a Registry API implementation.
# This includes the registry-server running with a local SQLite database,
# which can be started by running `registry-server -c config/sqlite.yaml`
# from the root of the registry repo. To configure clients to call this
# server, run `. auth/LOCAL.sh` in the shell before running the following
# commands.

# This script assumes that PROJECT is set to the name of your registry project.

# Get the commit hash of the checked-out protos directory
export COMMIT=`(cd ~/Desktop/googleapis; git rev-parse HEAD)`

# Upload all of the APIs in the googleapis directory at once.
# This happens in parallel and usually takes less than 10 seconds.
registry upload bulk protos \
	--project-id $PROJECT ~/Desktop/googleapis \
	--base-uri https://github.com/googleapis/googleapis/blob/$COMMIT 

# The `registry upload bulk protos` subcommand automatically generated API ids
# from the path to the protos in the repo. List the APIs with the following command:
registry list projects/$PROJECT/locations/global/apis

# We can count them by piping this through `wc -l`.
registry list projects/$PROJECT/locations/global/apis | wc -l

# Many of these APIs have multiple versions. We can list all of the API versions
# by using a "-" wildcard for the API id:
registry list projects/$PROJECT/locations/global/apis/-/versions

# Similarly, we can use wildcards for the version ids and list all of the specs.
# Here you'll see that the spec IDs are "protos.zip". This was set in the registry
# tool, which uploaded each API description as a zip archive of proto files.
registry list projects/$PROJECT/locations/global/apis/-/versions/-/specs

# To see more about an individual spec, use the `registry get` command:
registry get projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/google-cloud-translate-v3.zip

# You can also get this with direct calls to the registry rpc service:
registry rpc get-api-spec --name projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/google-cloud-translate-v3.zip

# Add the `--json` flag to get this as JSON:
registry rpc get-api-spec --name projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/google-cloud-translate-v3.zip --json

# You might notice that that didn't return the actual spec. That's because the spec contents
# are accessed through a separate method that (when transcoded to HTTP) allows direct download
# of spec contents.
registry rpc get-api-spec-contents --name projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/google-cloud-translate-v3.zip > protos-1.zip

# An easier way to get the bytes of the spec is to use `registry get` with the `--contents` flag.
# This writes the bytes to stdout, so you probably want to redirect this to a file, as follows:
registry get projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/google-cloud-translate-v3.zip \
	--contents > protos-2.zip

# When you unzip this file, you'll find a directory hierarchy suitable for compiling with `protoc`.
# protoc google/cloud/translate/v3/translation_service.proto -o.
# (This requires additional protos that you can find in
# [github.com/googleapis/api-common-protos](https://github.com/googleapis/api-common-protos).
