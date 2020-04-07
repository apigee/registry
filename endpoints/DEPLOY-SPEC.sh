#!/bin/bash

source ./CONFIG.sh

gcloud --verbosity=debug endpoints services deploy api_config.yaml proto.pb
