#!/bin/bash

source ./CONFIG.sh

gcloud --verbosity=debug endpoints services deploy api_config_open.yaml proto.pb
