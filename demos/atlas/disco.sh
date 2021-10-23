#!/bin/sh

# Upload all of the APIs from the Discovery Service at once.
# This happens in parallel and usually takes a minute or two.
registry upload bulk discovery \
        --project_id atlas

