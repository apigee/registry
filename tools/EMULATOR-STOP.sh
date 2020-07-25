#!/bin/bash

`gcloud beta emulators datastore env-unset`

pgrep -f 'gcloud.py beta emulators datastore start' | xargs kill -9
