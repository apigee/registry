#!/bin/sh

echo "Headers should include access-control-allow-origin\n"

curl $APG_REGISTRY_AUDIENCES/v1alpha1/status \
    -i \
    -H "Origin: *" \
    -X OPTIONS 