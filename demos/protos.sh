#!/bin/bash
#
# Copyright 2020 Google LLC. All Rights Reserved.
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
# Protocol Buffer descriptions of public Google APIs to build a colllection
# of API descriptions and them performs some analysis on them.
# It assumes that the repo has been cloned to the user's Desktop,
# so that it can be found at `~/Desktop/googleapis`.

# It also assumes an enviroment configured to call a Registry API implementation.
# This includes the registry-server running with a local SQLite database,
# which can be started by running `registry-server -c config/sqlite.yaml`
# from the root of the registry repo. To configure clients to call this
# server, run `source auth/LOCAL.sh` in the shell before running the following
# commands.

# First, delete the "protos" project to get a fresh start.
apg registry delete-project --name projects/protos

# Get the commit hash of the checked-out protos directory
export COMMIT=`(cd ~/Desktop/googleapis; git rev-parse HEAD)`

# Upload all of the APIs in the googleapis directory at once.
# This happens in parallel and usually takes less than 10 seconds.
registry upload bulk protos \
	--project_id protos ~/Desktop/googleapis \
	--base_uri https://github.com/googleapis/googleapis/blob/$COMMIT 

# The Atlas project was automatically created. Here we'll use an
# update-project call to set a few properties of the project.
apg registry update-project \
	--project.name "projects/protos" \
	--project.display_name "Protos" \
	--project.description "Protocol buffer descriptions of public Google APIs"

# Now compute summary details of all of the APIs in the project. 
# This will log errors if any of the API specs can't be parsed,
# but for every spec that is parsed, this will set the display name
# and description of the corresponding API from the values in the specs.
registry compute details projects/protos/apis/-

# The `registry upload bulk protos` subcommand automatically generated API ids
# from the path to the protos in the repo. List the APIs with the following command:
registry list projects/protos/apis

# We can count them by piping this through `wc -l`.
registry list projects/protos/apis | wc -l

# Many of these APIs have multiple versions. We can list all of the API versions
# by using a "-" wildcard for the API id:
registry list projects/protos/apis/-/versions

# Similarly, we can use wildcards for the version ids and list all of the specs.
# Here you'll see that the spec IDs are "protos.zip". This was set in the registry
# tool, which uploaded each API description as a zip archive of proto files.
registry list projects/protos/apis/-/versions/-/specs

# To see more about an individual spec, use the `registry get` command:
registry get projects/protos/apis/google-cloud-translate/versions/v3/specs/protos.zip

# You can also get this with the automatically-generated `apg` command line tool:
apg registry get-spec --name projects/protos/apis/google-cloud-translate/versions/v3/specs/protos.zip

# Add the `--json` flag to get this as JSON:
apg registry get-spec --name projects/protos/apis/google-cloud-translate/versions/v3/specs/protos.zip --json

# You might notice that this doesn't return the actual spec. That's because the get-spec
# API takes a `view` argument, and its default value ("BASIC") excludes the spec bytes.
# To get the spec contents, add "--view FULL" to your API call:
apg registry get-spec --name projects/protos/apis/google-cloud-translate/versions/v3/specs/protos.zip --json --view FULL

# An easier way to get the bytes of the spec is to use `registry get` with the `--contents` flag.
# This writes the bytes to stdout, so you probably want to redirect this to a file, as follows:
registry get projects/protos/apis/google-cloud-translate/versions/v3/specs/protos.zip --contents > protos.zip

# When you unzip this file, you'll find a directory hierarchy suitable for compiling with `protoc`.
# protoc google/cloud/translate/v3/translation_service.proto -o.
# (This requires additional protos that you can find in
# [github.com/googleapis/api-common-protos](https://github.com/googleapis/api-common-protos).

# The registry tool can compute simple complexity metrics for protos stored in the Registry.
registry compute complexity projects/protos/apis/-/versions/-/specs/-

# Complexity results are stored in properties associated with the specs.
registry list projects/protos/apis/-/versions/-/specs/-/properties/complexity

# We can use the `registry get` subcommand to read individual complexity records.
registry get projects/protos/apis/google-cloud-translate/versions/v3/specs/protos.zip/properties/complexity

# The registry tool also supports exporting all of the complexity results to a Google sheet.
# (The following command expects OAuth client credentials with access to the
# Google Sheets API to be available locally in ~/.credentials/registry.json)
registry export sheet projects/protos/apis/-/versions/-/specs/-/properties/complexity

# We can also compute the vocabulary of proto APIs.
registry compute vocabulary projects/protos/apis/-/versions/-/specs/-

# Vocabularies are also stored as properties associated with API specs.
registry get projects/protos/apis/google-cloud-translate/versions/v3/specs/protos.zip/properties/vocabulary

# The registry command can perform set operations on vocabularies.
# To find common terms in all Google speech-related APIs, use the following:
registry vocabulary intersection projects/protos/apis/-/versions/-/specs/-/properties/vocabulary --filter "api_id.contains('speech')"

# We can also save this to a property.
registry vocabulary intersection projects/protos/apis/-/versions/-/specs/-/properties/vocabulary --filter "api_id.contains('speech')" --output projects/protos/properties/speech-common

# We can then read it directly or export it to a Google Sheet.
registry get projects/protos/properties/speech-common
registry export sheet projects/protos/properties/speech-common

# To see a larger vocabulary, let's now compute the union of all the vocabularies in our project.
registry vocabulary union projects/protos/apis/-/versions/-/specs/-/properties/vocabulary --output projects/protos/properties/vocabulary

# We can also export this with `registry get` but it's easier to view this as a sheet:
registry export sheet projects/protos/properties/vocabulary

# You'll notice that usage counts are included for each term, so we can sort by count
# and find the most commonly-used terms across all of our APIs.
# With vocabulary operations we can discover common terms across groups of APIs,
# track changes across versions, and find unique terms in APIs that we are reviewing.
# By storing these results and other properties in the Registry, we can build a
# centralized store of API information that can help manage an API program.