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
# Protocol Buffer descriptions of public Google APIs to build a colllection
# of API descriptions and them performs some analysis on them.
# It assumes that the repo has been cloned to the user's Desktop,
# so that it can be found at `~/Desktop/googleapis`.

# It also assumes an enviroment configured to call a Registry API implementation.
# This includes the registry-server running with a local SQLite database,
# which can be started by running `registry-server -c config/sqlite.yaml`
# from the root of the registry repo. To configure clients to call this
# server, run `. auth/LOCAL.sh` in the shell before running the following
# commands.

# A registry exists under a top-level project.
PROJECT=protos

# First, delete and re-create the project to get a fresh start.
regctl rpc admin delete-project --name projects/$PROJECT
regctl rpc admin create-project --project_id $PROJECT \
	--project.display_name "Google APIs" \
	--project.description "Protocol buffer descriptions of public Google APIs"

# Get the commit hash of the checked-out protos directory
export COMMIT=`(cd ~/Desktop/googleapis; git rev-parse HEAD)`

# Upload all of the APIs in the googleapis directory at once.
# This happens in parallel and usually takes less than 10 seconds.
regctl upload bulk protos \
	--project-id $PROJECT ~/Desktop/googleapis \
	--base-uri https://github.com/googleapis/googleapis/blob/$COMMIT 

# The `regctl upload bulk protos` subcommand automatically generated API ids
# from the path to the protos in the repo. List the APIs with the following command:
regctl list projects/$PROJECT/locations/global/apis

# We can count them by piping this through `wc -l`.
regctl list projects/$PROJECT/locations/global/apis | wc -l

# Many of these APIs have multiple versions. We can list all of the API versions
# by using a "-" wildcard for the API id:
regctl list projects/$PROJECT/locations/global/apis/-/versions

# Similarly, we can use wildcards for the version ids and list all of the specs.
# Here you'll see that the spec IDs are "protos.zip". This was set in the registry
# tool, which uploaded each API description as a zip archive of proto files.
regctl list projects/$PROJECT/locations/global/apis/-/versions/-/specs

# To see more about an individual spec, use the `regctl get` command:
regctl get projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/google-cloud-translate-v3.zip

# You can also get this with the automatically-generated `regctl rpc` command line tool:
regctl rpc get-api-spec --name projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/google-cloud-translate-v3.zip

# Add the `--json` flag to get this as JSON:
regctl rpc get-api-spec --name projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/google-cloud-translate-v3.zip --json

# You might notice that that didn't return the actual spec. That's because the spec contents
# are accessed through a separate method that (when transcoded to HTTP) allows direct download
# of spec contents.
regctl rpc get-api-spec-contents --name projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/google-cloud-translate-v3.zip

# An easier way to get the bytes of the spec is to use `regctl get` with the `--contents` flag.
# This writes the bytes to stdout, so you probably want to redirect this to a file, as follows:
regctl get projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/google-cloud-translate-v3.zip \
	--contents > protos.zip

# When you unzip this file, you'll find a directory hierarchy suitable for compiling with `protoc`.
# protoc google/cloud/translate/v3/translation_service.proto -o.
# (This requires additional protos that you can find in
# [github.com/googleapis/api-common-protos](https://github.com/googleapis/api-common-protos).

# The regctl tool can compute simple complexity metrics for protos stored in the Registry.
regctl compute complexity projects/$PROJECT/locations/global/apis/-/versions/-/specs/-

# Complexity results are stored in artifacts associated with the specs.
regctl list projects/$PROJECT/locations/global/apis/-/versions/-/specs/-/artifacts/complexity

# We can use the `regctl get` subcommand to read individual complexity records.
regctl get projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/google-cloud-translate-v3.zip/artifacts/complexity

# The regctl tool also supports exporting all of the complexity results to a Google sheet.
# (The following command expects OAuth client credentials with access to the
# Google Sheets API to be available locally in ~/.credentials/registry.json)
regctl export sheet projects/$PROJECT/locations/global/apis/-/versions/-/specs/-/artifacts/complexity \
	--as projects/$PROJECT/locations/global/artifacts/complexity-sheet

# We can also compute the vocabulary of proto APIs.
regctl compute vocabulary projects/$PROJECT/locations/global/apis/-/versions/-/specs/-

# Vocabularies are also stored as artifacts associated with API specs.
regctl get projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/google-cloud-translate-v3.zip/artifacts/vocabulary

# The regctl command can perform set operations on vocabularies.
# To find common terms in all Google speech-related APIs, use the following:
regctl vocabulary intersection projects/$PROJECT/locations/global/apis/-/versions/-/specs/-/artifacts/vocabulary \
	--filter "api_id.contains('speech')"

# We can also save this to a property.
regctl vocabulary intersection projects/$PROJECT/locations/global/apis/-/versions/-/specs/-/artifacts/vocabulary \
	--filter "api_id.contains('speech')" --output projects/$PROJECT/locations/global/artifacts/speech-common

# We can then read it directly or export it to a Google Sheet.
regctl get projects/$PROJECT/locations/global/artifacts/speech-common
regctl export sheet projects/$PROJECT/locations/global/artifacts/speech-common

# To see a larger vocabulary, let's now compute the union of all the vocabularies in our project.
regctl vocabulary union projects/$PROJECT/locations/global/apis/-/versions/-/specs/-/artifacts/vocabulary \
	--output projects/$PROJECT/locations/global/artifacts/vocabulary

# We can also export this with `regctl get` but it's easier to view this as a sheet:
regctl export sheet projects/$PROJECT/locations/global/artifacts/vocabulary

# You'll notice that usage counts are included for each term, so we can sort by count
# and find the most commonly-used terms across all of our APIs.
# With vocabulary operations we can discover common terms across groups of APIs,
# track changes across versions, and find unique terms in APIs that we are reviewing.
# By storing these results and other artifacts in the Registry, we can build a
# centralized store of API information that can help manage an API program.

# We can also run analysis tools like linters and store the results in the Registry.
# Here we run the Google api-linter and compile summary statistics.
regctl compute lint projects/$PROJECT/locations/global/apis/-/versions/-/specs/-
regctl compute lintstats projects/$PROJECT/locations/global/apis/-/versions/-/specs/- --linter aip
regctl compute lintstats projects/$PROJECT --linter aip