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

# This script uses the [OpenAPI Directory](https://github.com/APIs-guru/openapi-directory)
# to build a collection of APIs and then performs some analysis on the uploaded API
# descriptions. It assumes that this repo has been cloned to the user's Desktop,
# so that it can be found at `~/Desktop/openapi-directory`.

# It also assumes an environment configured to call a Registry API implementation.
# This includes the registry-server running with a local SQLite database,
# which can be started by running `registry-server -c config/sqlite.yaml`
# from the root of the registry repo. To configure clients to call this
# server, run `source auth/LOCAL.sh` in the shell before running the following
# commands.

# A registry exists under a top-level project.
PROJECT=openapi

# First, delete and re-create the "openapis" project to get a fresh start.
apg admin delete-project --name projects/$PROJECT
apg admin create-project --project_id $PROJECT \
	--project.display_name "OpenAPI Directory" \
	--project.description "APIs collected from the APIs.guru OpenAPI Directory"

# Get the commit hash of the checked-out OpenAPI directory
export COMMIT=`(cd ~/Desktop/openapi-directory; git rev-parse HEAD)`

# Upload all of the APIs in the OpenAPI directory at once.
# This happens in parallel and usually takes around 2 minutes.
registry upload bulk openapi \
	--project-id $PROJECT ~/Desktop/openapi-directory/APIs \
	--base-uri https://github.com/APIs-guru/openapi-directory/blob/$COMMIT/APIs 

# Now compute summary details of all of the APIs in the project. 
# This will log errors if any of the API specs can't be parsed,
# but for every spec that is parsed, this will set the display name
# and description of the corresponding API from the values in the specs.
registry compute details projects/$PROJECT/locations/global/apis/-

# Verify this for one of the APIs.
registry get projects/$PROJECT/locations/global/apis/wordnik.com

# You can also get this API with the lower-level apg command.
# Add the --json option to get JSON-formatted output.
apg registry get-api --name projects/$PROJECT/locations/global/apis/wordnik.com --json

# Get the API spec
apg registry get-api-spec --name projects/$PROJECT/locations/global/apis/wordnik.com/versions/4.0/specs/openapi.yaml

# You might notice that that didn't return the actual spec. That's because the spec contents
# are accessed through a separate method that (when transcoded to HTTP) allows direct download
# of spec contents.
apg registry get-api-spec-contents --name projects/$PROJECT/locations/global/apis/wordnik.com/versions/4.0/specs/openapi.yaml

# If you have jq and the base64 tool installed, you can get the spec contents from the RPC response.
# apg registry get-api-spec-contents --name projects/$PROJECT/locations/global/apis/wordnik.com/versions/4.0/specs/openapi.yaml --json | jq .data -r | base64 --decode

# Another way to get the bytes of the spec is to use `registry get` with the `--contents` flag.
registry get projects/$PROJECT/locations/global/apis/wordnik.com --contents

# List all of the APIs in the project.
registry list projects/$PROJECT/locations/global/apis

# List all of the versions of an API.
registry list projects/$PROJECT/locations/global/apis/wordnik.com/versions

# List all of the specs associated with an API version.
registry list projects/$PROJECT/locations/global/apis/wordnik.com/versions/4.0/specs

# Following [AIP-159](https://google.aip.dev/159), the list operations support the "-" wildcard.
# This allows us to list objects across multiple collections.
registry list projects/$PROJECT/locations/global/apis/wordnik.com/versions/-/specs/-

# Now let's list all of the specs in the project.
registry list projects/$PROJECT/locations/global/apis/-/versions/-/specs/-

# That's a lot. Let's count them with `wc -l`.
registry list projects/$PROJECT/locations/global/apis/-/versions/-/specs/- | wc -l

# Using wildcards, we can list all of the specs with a particular version.
registry list projects/$PROJECT/locations/global/apis/-/versions/1.0.0/specs/-

# List operations also support filtering by following [AIP-160](https://google.aip.dev/160).
# Filter functions are evaluated using [CEL](https://github.com/google/cel-spec).
# Here's an example:
registry list projects/$PROJECT/locations/global/apis/-/versions/-/specs/- \
  --filter "api_id.startsWith('goog')"

# This is a bit more verbose than glob expressions but much more powerful.
# You can also refer to other fields in the messages that match the pattern:
registry list projects/$PROJECT/locations/global/apis/- --filter "description.contains('speech')"

# The registry command can also compute some basic API properties.
# This computes simple complexity metrics for every spec in the project.
registry compute complexity projects/$PROJECT/locations/global/apis/-/versions/-/specs/-

# The complexity metrics are stored in artifacts associated with each spec.
# In this case, the artifacts were stored in a serialized protocol buffer.
# You can get their values with the "get" subcommand.
registry get projects/$PROJECT/locations/global/apis/wordnik.com/versions/4.0/specs/openapi.yaml/artifacts/complexity --contents

# It's also possible to export artifacts to a Google sheet.
# (The following command expects OAuth client credentials with access to the
# Google Sheets API to be available locally in ~/.credentials/registry.json)
registry export sheet projects/$PROJECT/locations/global/apis/-/versions/-/specs/-/artifacts/complexity \
	--as projects/$PROJECT/locations/global/artifacts/complexity-sheet

# Another interesting property that can be computed is the "vocabulary" of an API.
# The following command computes vocabularies of every API spec in the project.
registry compute vocabulary projects/$PROJECT/locations/global/apis/-/versions/-/specs/-

# Vocabularies are stored in the "vocabulary" property.
registry get projects/$PROJECT/locations/global/apis/wordnik.com/versions/4.0/specs/openapi.yaml/artifacts/vocabulary --contents

# The registry command can perform set operations on vocabularies.
# To find common terms in all Google APIs, use the following:
registry vocabulary intersection projects/$PROJECT/locations/global/apis/-/versions/-/specs/-/artifacts/vocabulary \
  --filter "api_id.startsWith('googleapis')"

# We can also save this to a property.
registry vocabulary intersection projects/$PROJECT/locations/global/apis/-/versions/-/specs/-/artifacts/vocabulary \
  --filter "api_id.startsWith('googleapis')" \
  --output projects/$PROJECT/locations/global/artifacts/google-common

# We can then read it directly or export it to a Google Sheet.
registry get projects/$PROJECT/locations/global/artifacts/google-common
registry export sheet projects/$PROJECT/locations/global/artifacts/google-common

# With vocabulary operations we can discover common terms across groups of APIs,
# track changes across versions, and find unique terms in APIs that we are reviewing.
# By storing these results and other artifacts in the Registry, we can build a
# centralized store of API information that can help manage an API program.
