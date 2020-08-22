#!/bin/sh

curl $APG_REGISTRY_AUDIENCES/v1alpha1/projects/demo/apis -i -H "Authorization: Bearer $APG_REGISTRY_TOKEN"

