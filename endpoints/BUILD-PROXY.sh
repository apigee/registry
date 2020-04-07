#!/bin/bash

source ./CONFIG.sh

./gcloud_build_image.sh \
    -s $CLOUD_RUN_HOSTNAME \
    -c $CONFIG_ID \
    -p $ESP_PROJECT_ID