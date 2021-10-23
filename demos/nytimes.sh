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

# First, delete the "nytimes" project to get a fresh start.
apg registry delete-project --name projects/nytimes

# Get the commit hash of the checked-out OpenAPI directory
export COMMIT=`(cd ~/Desktop/openapi-directory; git rev-parse HEAD)`

# Upload all of the APIs in the OpenAPI directory.
registry upload bulk openapi \
	--project_id nytimes ~/Desktop/openapi-directory/APIs/nytimes.com \
	--base_uri https://github.com/APIs-guru/openapi-directory/blob/$COMMIT/APIs/nytimes.com 

# The nytimes project was automatically created. Here we'll use an
# update-project call to set a few properties of the project.
apg registry update-project \
	--project.name "projects/nytimes" \
	--project.display_name "New York Times APIs" \
	--project.description "APIs collected from the APIs.guru OpenAPI Directory"

# Now compute summary details of all of the APIs in the project. 
# This will log errors if any of the API specs can't be parsed,
# but for every spec that is parsed, this will set the display name
# and description of the corresponding API from the values in the specs.
registry compute details projects/nytimes/apis/-

# Verify this for one of the APIs.
registry get projects/nytimes/apis/archive

# You can also get this API with the lower-level apg command.
# Add the --json option to get JSON-formatted output.
apg registry get-api --name projects/nytimes/apis/archive --json

# List all of the APIs in the project.
registry list projects/nytimes/apis

# List all of the versions of an API.f
registry list projects/nytimes/apis/archive/versions

# List all of the specs associated with an API version.
registry list projects/nytimes/apis/archive/versions/1.0.0/specs

# Following [AIP-159](https://google.aip.dev/159), the list operations support the "-" wildcard.
# This allows us to list objects across multiple collections.
registry list projects/nytimes/apis/archive/versions/-/specs/-

# Now let's list all of the specs in the project.
registry list projects/nytimes/apis/-/versions/-/specs/-

# Let's count them with `wc -l`.
registry list projects/nytimes/apis/-/versions/-/specs/- | wc -l

# Using wildcards, we can list all of the specs with a particular version.
registry list projects/nytimes/apis/-/versions/1.0.0/specs/-

# List operations also support filtering by following [AIP-160](https://google.aip.dev/160).
# Filter functions are evaluated using [CEL](https://github.com/google/cel-spec).
# Here's an example:
registry list projects/nytimes/apis/-/versions/-/specs/- --filter "api_id.startsWith('books')"

# This is a bit more verbose than glob expressions but much more powerful.
# You can also refer to other fields in the messages that match the pattern:
registry list projects/nytimes/apis/- --filter "description.contains('articles')"

# The registry command can also compute some basic API properties.
# This computes simple complexity metrics for every spec in the project.
registry compute complexity projects/nytimes/apis/-/versions/-/specs/-

# The complexity metrics are stored in artifacts associated with each spec.
# In this case, the artifacts were stored in a serialized protocol buffer.
# You can get their values with the "get" subcommand.
registry get projects/nytimes/apis/archive/versions/1.0.0/specs/swagger.yaml/artifacts/complexity

# It's also possible to export artifacts to a Google sheet.
# (The following command expects OAuth client credentials with access to the
# Google Sheets API to be available locally in ~/.credentials/registry.json)
registry export sheet projects/nytimes/apis/-/versions/-/specs/-/artifacts/complexity \
	--as projects/nytimes/artifacts/complexity-sheet

# Another interesting property that can be computed is the "vocabulary" of an API.
# The following command computes vocabularies of every API spec in the project.
registry compute vocabulary projects/nytimes/apis/-/versions/-/specs/-

# Vocabularies are stored in the "vocabulary" property.
registry get projects/nytimes/apis/archive/versions/1.0.0/specs/swagger.yaml/artifacts/vocabulary

# The registry command can perform set operations on vocabularies.
# To find all terms in all APIs, use the following:
registry vocabulary union projects/nytimes/apis/-/versions/-/specs/-/artifacts/vocabulary

# We can also save this to a property.
registry vocabulary union projects/nytimes/apis/-/versions/-/specs/-/artifacts/vocabulary --output projects/nytimes/artifacts/vocabulary-all

# We can then read it directly or export it to a Google Sheet.
registry get projects/nytimes/artifacts/vocabulary-all
registry export sheet projects/nytimes/artifacts/vocabulary-all

# With vocabulary operations we can discover common terms across groups of APIs,
# track changes across versions, and find unique terms in APIs that we are reviewing.
# By storing these results and other artifacts in the Registry, we can build a
# centralized store of API information that can help manage an API program.

# We can also run analysis tools like linters and store the results in the Registry.
# Here we run the Spectral linter and compile summary statistics.
registry compute lint projects/nytimes/apis/-/versions/-/specs/- --linter spectral
registry compute lintstats projects/nytimes/apis/-/versions/-/specs/- --linter spectral
registry compute lintstats projects/nytimes --linter spectral

