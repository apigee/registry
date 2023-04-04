# gke

This directory contains configuration tools and other support files for running
`registry-server` on GKE with envoy setup.

## Instructions

Following steps assume you're in the root directory.

1. Run `make build` to build the docker image and upload to GCR. The default
   server configuration is
   [registry-server.yaml](../../config/registry-server.yaml). You can provide a
   different configuration by replacing this file or by setting configuration
   using environment variables.

1. Create a GKE deployment and expose the backend server through a load
   balancer:

   - To create an external LB, run `make deploy-gke`
   - To create an internal LB, run `make deploy-gke LB=internal`

1. Setup the client authentication. This step differs based on the load balancer
   type you chose in the previous step:

   - External LB: run `. auth/GKE.sh`.
   - Internal LB: Usually you can't access services that are behind the internal
     LB from your local. For more details, please check
     [here](https://cloud.google.com/kubernetes-engine/docs/how-to/internal-load-balancing#inspect).

     1. Find the ingress IP:
        `kubectl get service registry-backend -o jsonpath="{.status.loadBalancer.ingress[0].ip}"`.
     1. Find the service port:
        `kubectl get service registry-backend -o jsonpath="{.spec.ports[0].port}"`
     1. SSH into a VM.
     1. Run the following commands with `ingress_ip` found in the first step and
        `service_port` found in the second step:

        ```shell script
        export REGISTRY_ADDRESS="<ingress_ip>:<service_port>"
        export REGISTRY_CLIENT_EMAIL=$(gcloud config list account --format "value(core.account)")
        export REGISTRY_TOKEN=$(gcloud auth print-identity-token ${REGISTRY_CLIENT_EMAIL})
        ```

1. Verify the server. The GKE deployment uses
   `<PROJECT_NUMBER>-compute@developer.gserviceaccount.com` by default. Please
   ensure the service account has sufficient permissions to access the database
   you configured. Below is a sample curl call to access your GKE deployment:

   ```shell script
   curl $REGISTRY_ADDRESS/v1/status -i -H "Authorization: Bearer $REGISTRY_TOKEN"
   ```
