#!/bin/bash
#
# Copyright 2021 Google LLC.
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

# This script uses the [OpenAPI Directory](https://github.com/APIs-guru/openapi-directory)
# to build a collection of APIs and then performs some analysis on the uploaded API
# descriptions. It assumes that this repo has been cloned to the user's Desktop,
# so that it can be found at `~/Desktop/openapi-directory`.

# It also assumes an environment configured to call a Registry API implementation.
# This includes the registry-server running with a local SQLite database,
# which can be started by running `registry-server -c config/sqlite.yaml`
# from the root of the registry repo. To configure clients to call this
# server, run `. auth/LOCAL.sh` in the shell before running the following
# commands.

# This script assumes that PROJECT is set to the name of your registry project.

# Get the commit hash of the checked-out OpenAPI directory
export COMMIT=`(cd ~/Desktop/openapi-directory; git rev-parse HEAD)`

# Upload all of the APIs in the OpenAPI directory at once.
# This happens in parallel and usually takes around 2 minutes.
registry upload openapi \
	--project-id $PROJECT ~/Desktop/openapi-directory/APIs \
	--base-uri https://github.com/APIs-guru/openapi-directory/blob/$COMMIT/APIs 

# Get one of the APIs.
registry get projects/$PROJECT/locations/global/apis/wordnik.com

# You can also get this API with direct calls to the registry rpc service.
# Add the --json option to get JSON-formatted output.
registry rpc get-api --name projects/$PROJECT/locations/global/apis/wordnik.com --json

# Get the API spec
registry rpc get-api-spec --name projects/$PROJECT/locations/global/apis/wordnik.com/versions/4.0/specs/openapi

# You might notice that didn't return the actual spec. That's because the spec contents
# are accessed through a separate method that (when transcoded to HTTP) allows direct download
# of spec contents.
registry rpc get-api-spec-contents --name projects/$PROJECT/locations/global/apis/wordnik.com/versions/4.0/specs/openapi

# If you have jq and the base64 tool installed, you can get the spec contents from the RPC response.
# registry rpc get-api-spec-contents --name projects/$PROJECT/locations/global/apis/wordnik.com/versions/4.0/specs/openapi --json | jq .data -r | base64 --decode

# Another way to get the bytes of the spec is to use `registry get` with the `--contents` flag.
registry get projects/$PROJECT/locations/global/apis/wordnik.com --contents

# List all of the APIs in the project.
registry get projects/$PROJECT/locations/global/apis

# List all of the versions of an API.
registry get projects/$PROJECT/locations/global/apis/wordnik.com/versions

# List all of the specs associated with an API version.
registry get projects/$PROJECT/locations/global/apis/wordnik.com/versions/4.0/specs

# Following [AIP-159](https://google.aip.dev/159), the list operations support the "-" wildcard.
# This allows us to list objects across multiple collections.
registry get projects/$PROJECT/locations/global/apis/wordnik.com/versions/-/specs/-

# Now let's list all of the specs in the project.
registry get projects/$PROJECT/locations/global/apis/-/versions/-/specs/-

# That's a lot. Let's count them with `wc -l`.
registry get projects/$PROJECT/locations/global/apis/-/versions/-/specs/- | wc -l

# Using wildcards, we can list all of the specs with a particular version.
registry get projects/$PROJECT/locations/global/apis/-/versions/1.0.0/specs/-

# List operations also support filtering by following [AIP-160](https://google.aip.dev/160).
# Filter functions are evaluated using [CEL](https://github.com/google/cel-spec).
# Here's an example:
registry get projects/$PROJECT/locations/global/apis/-/versions/-/specs/- \
  --filter "api_id.startsWith('goog')"

# This is a bit more verbose than glob expressions but much more powerful.
# You can also refer to other fields in the messages that match the pattern:
registry get projects/$PROJECT/locations/global/apis/- --filter "description.contains('speech')"
