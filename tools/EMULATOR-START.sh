#!/bin/bash

# https://cloud.google.com/sdk/gcloud/reference/beta/emulators/datastore/start

gcloud components install beta --quiet
gcloud components install cloud-datastore-emulator --quiet
gcloud beta emulators datastore start --no-store-on-disk &
`gcloud beta emulators datastore env-init`
