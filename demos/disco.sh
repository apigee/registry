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

# This script uses the [Google API Discovery Service](https://developers.google.com/discovery)
# to build a collection of Google API descriptions and then performs some analysis on these
# discovery documents. Because it calls the Discovery Service directly, no files need to be
# downloaded to the user's local system.

# The script assumes an environment configured to call a Registry API implementation.
# This includes the registry-server running with a local SQLite database,
# which can be started by running `registry-server -c config/sqlite.yaml`
# from the root of the registry repo. To configure clients to call this
# server, run `. auth/LOCAL.sh` in the shell before running the following
# commands.

# This script assumes that PROJECT is set to the name of your registry project.

# Upload all of the APIs from the Discovery Service at once.
# This happens in parallel and usually takes a minute or two.
registry upload discovery \
	--project-id $PROJECT

# We can get a list of APIs with the following command:
registry get projects/$PROJECT/locations/global/apis

# We can count the APIs by piping this through `wc -l`.
registry get projects/$PROJECT/locations/global/apis | wc -l

# Many of these APIs have multiple versions. We can list all of the API versions
# by using a "-" wildcard for the API id:
registry get projects/$PROJECT/locations/global/apis/-/versions

# Similarly, we can use wildcards for the version ids and list all of the specs.
# Here you'll see that the spec IDs are "discovery". This was set in the registry
# tool, which uploaded each API description as gzipped JSON.
registry get projects/$PROJECT/locations/global/apis/-/versions/-/specs

# To see more about an individual spec, use the `-o yaml` option:
registry get projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/discovery -o yaml

# You can also get this with direct calls to the registry rpc service:
registry rpc get-api-spec \
	--name projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/discovery

# Add the `--json` flag to get this as JSON:
registry rpc get-api-spec --json \
	--name projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/discovery

# You might notice that didn't return the actual spec. That's because the spec contents
# are accessed through a separate method that (when transcoded to HTTP) allows direct download
# of spec contents.
registry rpc get-api-spec-contents \
	--name projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/discovery

# Another way to get the bytes of the spec is to use `registry get` with the `-o contents` flag.
registry get projects/$PROJECT/locations/global/apis/translate/versions/v3/specs/discovery \
	-o contents
