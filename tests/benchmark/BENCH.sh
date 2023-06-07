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

export PROJECTID=bench

# Assume that "bench" is a local project and all other names are hosted.
if [[ "$PROJECTID" == "bench" ]]; then
  . auth/LOCAL.sh
  # Redirect the following command to /dev/null to ignore errors due to nonexistent projects.
  registry rpc admin delete-project --name=projects/$PROJECTID --force --json &> /dev/null
  registry rpc admin create-project --project_id=$PROJECTID --json
else
  . auth/HOSTED.sh
fi

# Increase this to get better sampling.
export ITERATIONS=1

go test ./tests/benchmark -parallel=1 --bench=. --project_id=$PROJECTID --benchtime=${ITERATIONS}x --timeout=0
