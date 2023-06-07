#!/bin/bash
#
# Copyright 2022 Google LLC.
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

# A registry exists under a top-level project.
# Set the PROJECT environment variable to the name of your project.
PROJECT=sample

# Self-hosted (open source) installations require project creation.

# Delete and re-create the project to get a fresh start.
registry rpc admin delete-project --name projects/$PROJECT --force >& /dev/null
registry rpc admin create-project --project_id $PROJECT \
	--project.display_name $PROJECT \
	--project.description "A registry project" \
	--json

