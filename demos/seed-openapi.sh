#!/bin/bash

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