This directory contains code for deploying the worker architecture. The
architecture includes the following things:

- Disptacher: GKE workload listening to
  [Cloud Pub/Sub](https://cloud.google.com/pubsub) and creating tasks in a
  [Cloud Tasks](https://cloud.google.com/tasks) Queue.
- Cloud Tasks Queue: Responsible for triggering tasks schedules in the Queue
  (Trigger happen by invoking the worker endpoints).
- Workers: These are GKE workloads which are exposed as HTTP endpoints. Workers
  are responsible for executing specific tasks.

The current implementation includes and example computing lint results
automatically everytime a new spec in created/updated in the registry.

# Deployment:

Note: Please execute the following instructions from the root directory of this
project.

## Prerequisites:

- Set the env variable `REGISTRY_PROJECT_IDENTIIFIER` to correct project ID.
- This doc assumes that the registry application is already deployed on GKE.
  You can do that using the following:
  ```
  make build
  make deploy-gke
  ```
  **Note:** Make sure when you are deploying the registry application, you have
  `notify: true` set in the registry config file.

## Deploying worker setup:

- Once you have the registry up and running, let's deploy the supporting worker
  architecture.
- Build the worker and the dispatcher images:
  ```
  make build-workers
  ```
  This uses Cloud Build to build the images and stores them in Google Container
  Registry.
- Deploy the architecture:
  ```
  make deploy-workers
  ```
  This will deploy the whole architecture.

## Verification:

- To verify that your set-up is working, you can try to create a spec in the
  registry and check if the lint is getting computed automatically.
  - Set the local auth token:
    ```
    source auth/GKE.sh
    ```
  - Follow the script `demo.sh` to verify the functionality.

## Troubleshooting:

- Get the pods:

  ```
  kubectl get pods

  NAME                                  READY   STATUS    RESTARTS   AGE
  registry-backend-655479b6bb-h2mxv     1/1     Running   0          21h
  registry-dispatcher-564dbc584-v7mh8   1/1     Running   0          21h
  registry-worker-9dbd76c78-22dqk       1/1     Running   0          18h
  ```

- Get logs:
  ```
  kubectl logs registry-dispatcher-564dbc584-v7mh8
  kubectl logs registry-worker-9dbd76c78-22dqk
  ```
