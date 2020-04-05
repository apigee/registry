#!/bin.sh

source ./CONFIG.sh

gcloud run deploy $CLOUD_RUN_SERVICE_NAME \
  --image="gcr.io/$ESP_PROJECT_ID/endpoints-runtime-serverless:$CLOUD_RUN_HOSTNAME-$CONFIG_ID" \
  --allow-unauthenticated \
  --platform managed \
  --project=$ESP_PROJECT_ID