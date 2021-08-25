This directory contains all the configurations for the controller framework. It includes the following two things:
* A GKE  cron job which periodically executes the registry resolve command to update the state of the registry.
* A visualization dashboard which will show hoe many tasks were computed by the controller job in each execution.

* Controller job:
  - The `gke-job` directory contains the config for a GKE cron job.
  - Make sure you have uploaded a manifest to the registry first before deploying the cron job.
  Example: `registry upload manifest cmd/registry/controller/testdata/manifest_e2e.yaml --project_id=demo`
  - Once the manifets is uploaded, deploy the cron job
  	`make deploy-controller-job`

* Controller dashboard:
  - The `dasboard` directory includes two metrics which track the execution and task generation carried out by the controller job, and a daashboard which is built on top of these two metrics.
  To create the dashboard:
  `make deploy-controller-dashboard`


It is also possible to deploy the whole setup through a singular make target:
`make deploy-controller`