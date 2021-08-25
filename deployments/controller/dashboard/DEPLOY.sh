# Copyright 2021 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

metric=$( gcloud logging metrics list --filter=name=task_execution --format="value(name)" )
if [[ $metric ]]; then
	echo "Metric task_execution already exists"
else
	echo "Creating new metric"
	envsubst < deployments/controller/dashboard/metrics/task_execution.yaml | gcloud logging metrics create task_execution --config-from-file=-
fi

metric=$( gcloud logging metrics list --filter=name=task_generation --format="value(name)" )
if [[ $metric ]]; then
        echo "Metric task_generation already exists"
else
        echo "Creating new metric"
	envsubst < deployments/controller/dashboard/metrics/task_generation.yaml | gcloud logging metrics create task_generation --config-from-file=-
fi

project_no=$( gcloud projects describe $REGISTRY_PROJECT_IDENTIFIER --format="value(projectNumber)" )
dashboard_id=projects/$project_no/dashboards/registry-controller-status
dashboard=$( gcloud monitoring dashboards list --filter=name=$dashboard_id)

if [[ $dashboard ]]; then
	echo "Dashboard $dashboard_id already exists"
else
	echo "Creating new dashboard"
	envsubst < deployments/controller/dashboard/chart.yaml |  gcloud monitoring dashboards create --config-from-file=-
fi
