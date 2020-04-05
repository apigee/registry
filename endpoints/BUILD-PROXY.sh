#!/bin/sh

CLOUD_RUN_HOSTNAME="flame-endpoint-ozfrf5bp4a-uw.a.run.app"
CONFIG_ID="2020-04-03r293"
ESP_PROJECT_ID="your-project-identifier"

./gcloud_build_image.sh \
    -s $CLOUD_RUN_HOSTNAME \
    -c $CONFIG_ID \
    -p $ESP_PROJECT_ID