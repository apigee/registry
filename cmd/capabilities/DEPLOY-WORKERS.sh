set -e

gcloud config set project ${REGISTRY_PROJECT_IDENTIFIER}

# Get credentials for kubectl
gcloud container clusters get-credentials registry-backend --zone us-central1-a

registry_ingress_ip=$(kubectl get service registry-backend -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
registry_service_port=$(kubectl get service registry-backend -o jsonpath="{.spec.ports[0].port}")
if [ -z "${registry_ingress_ip}" ]; then
  echo "Please make sure the Registry service is deployed to GKE and exposed externally."; exit 1
fi

export APG_REGISTRY_ADDRESS="${registry_ingress_ip}:${registry_service_port}"

project_number=$(gcloud projects describe ${REGISTRY_PROJECT_IDENTIFIER} --format="value(projectNumber)")

# Set correct IAM permissions to default SAs
gcloud projects add-iam-policy-binding ${REGISTRY_PROJECT_IDENTIFIER} \
       --member="serviceAccount:${project_number}-compute@developer.gserviceaccount.com" \
       --role=roles/cloudtasks.enqueuer

gcloud projects add-iam-policy-binding ${REGISTRY_PROJECT_IDENTIFIER} \
       --member="serviceAccount:${project_number}-compute@developer.gserviceaccount.com" \
       --role=roles/pubsub.subscriber

# Registry worker
deployment=$( kubectl get deployment registry-worker || true)
if [[ $deployment ]]; then
    echo "Deleting existing deployment for registry-worker ..."
    envsubst < cmd/capabilities/worker/worker-deployment.yaml | kubectl apply -f -
fi

echo "Creating new deployment for registry-worker"
envsubst < cmd/capabilities/worker/worker-deployment.yaml | kubectl apply -f -

# Worker service
kubectl apply -f cmd/capabilities/worker/service.yaml
echo "Sleeping for registry-worker-service to get provisioned"
worker_ingress_ip=$(kubectl get service registry-worker-service -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
while [ ! $worker_ingress_ip ]
do
  sleep 5s
  worker_ingress_ip=$(kubectl get service registry-worker-service -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
done
worker_port=$(kubectl get service registry-worker-service -o jsonpath="{.spec.ports[0].port}")
export WORKER_URL="http://${worker_ingress_ip}:${worker_port}"

# Setup App Engine app for Cloud Tasks
app=$( gcloud app describe || true )
if [[ $app ]]; then
  echo "App Engine is already enabled."
else
  gcloud app create --region=us-central
fi

# Create Cloud Tasks Queue
task_queue=$( gcloud tasks queues describe registry-task-queue || true )
if [[ $task_queue ]]; then
  echo "Task Queue already exists."
else
  gcloud tasks queues create registry-task-queue
fi

export TASK_QUEUE_ID=$(gcloud tasks queues describe registry-task-queue --format "value(name)")

# Registry Deployment
deployment=$( kubectl get deployment registry-dispatcher || true)
if [[ $deployment ]]; then
    echo "Deleting existing deployment for registry-dispatcher ..."
    envsubst < cmd/capabilities/dispatcher/dispatcher-deployment.yaml | kubectl apply -f -
fi

envsubst < cmd/capabilities/dispatcher/dispatcher-deployment.yaml | kubectl apply -f -

