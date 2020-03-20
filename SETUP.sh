#!/bin/sh

#gcloud auth login

export CLI_FLAME_CLIENT_EMAIL="flame-client@your-project-identifier.iam.gserviceaccount.com"

export CLI_FLAME_AUDIENCES=$(gcloud beta run services describe flame --platform managed --format="value(status.address.url)")

export CLI_FLAME_ADDRESS=${CLI_FLAME_AUDIENCES#https://}:443

export CLI_FLAME_TOKEN=$(gcloud auth print-identity-token ${CLI_FLAME_CLIENT_EMAIL} --audiences="$CLI_FLAME_AUDIENCES")
