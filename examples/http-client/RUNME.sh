#!/bin/sh

curl $APG_REGISTRY_AUDIENCES/v1alpha1/status \
	-i \
	-H "Authorization: Bearer $APG_REGISTRY_TOKEN"

