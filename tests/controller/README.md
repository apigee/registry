## Registry Controller

In simple terms, controller is a framework to keep the registry up-to-date. The idea for this framework is inspired from the [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/) pattern. Controller can be used to enrich the API data stored in the registry, with additional analytical artifacts. 

### Configuration details
A Controller will take in configuration (aka manifest) which describes the desired state of the registry. The desired state in this context is essentially dependency relations between different entities in the database. The controller’s job is to keep these dependency relations valid and up-to-date. It will take actions to eventually get the registry to its desired state.

Example manifest:
```
id: "example-manifest"
generated_resources:
- pattern: apis/-/versions/-/specs/-/artifacts/lint-spectral
  dependencies:
    - pattern: $resource.spec
      filter: "mime_type.contains('openapi')"
  action: "registry compute lint $resource.spec --linter spectral"
```
The above manifest defines that for every spec (`pattern: $resource.spec`) of type openapi (`filter: "mime_type.contains('openapi')"`) in the registry, there should exist a corresponding lint-spectral artifact (`pattern: apis/-/versions/-/specs/-/artifacts/lint-spectral`). If the target artifact is missing or out-of-date, then it should be re-generated using the specified action (`action: "registry compute lint $resource.spec --linter spectral"`).

The controller will compare the current and the supplied desired state of the registry. For resources which are non-existing or outdated, the controller will execute the action supplied in the manifest.

### Usage

With this basic definition in mind, you can supply different configurations to the controller to generate various artifacts in the registry.

#### Setup the registry server
Make sure you have a registry server running. We will deploy the server to GKE for the purpose of this walkthrough. 

- Deploy the registry server to GKE. The following two make targets will build and deploy the server:
  ```shell
  # Build
  make build
  
  # Deploy
  make deploy-gke
  ``` 
  Follow these [instructions](https://github.com/apigee/registry/blob/main/README.md#running-the-registry-api-server-on-gke) for more details.

- Populate some resources in the registry:
  ```shell
  # Setup auth
  source auth/GKE.sh

  # Verify the server is running
  apg admin get-status

  # Create demo project
  apg admin create-project --project_id demo --json

  # Create 10 versions of the petstore API
  ./tests/controller/create_apis.sh 10
  ``` 

#### Upload the manifest
```shell
registry upload manifest tests/controller/testdata/manifest.yaml --project-id=demo

```
The contents of the manifest that we are uploading is as follows:
```yaml
id: "test-manifest"
generated_resources:
  - pattern: "apis/-/versions/-/specs/-/artifacts/lint-spectral"
    dependencies:
      - pattern: "$resource.spec"
        filter: "mime_type.contains('openapi')"
    action: "registry compute lint $resource.spec --linter spectral"
  - pattern: "apis/-/versions/-/specs/-/artifacts/lintstats-spectral"
    dependencies:
      - pattern: "$resource.spec/artifacts/lint-spectral"
      - pattern: "$resource.spec/artifacts/complexity"
    action: "registry compute lintstats $resource.spec --linter spectral"
  - pattern: "apis/-/versions/-/specs/-/artifacts/vocabulary"
    dependencies:
      - pattern: "$resource.spec"
    action: "registry compute vocabulary $resource.spec"
  - pattern: "apis/-/versions/-/specs/-/artifacts/complexity"
    dependencies:
      - pattern: "$resource.spec"
    action: "registry compute complexity $resource.spec"
```

#### Run the controller
There are three ways to run the controller:

##### Dry-run mode:
In this mode, the controller will generate the actions, which it calculates, are needed to bring registry up-to-date. However, it will not execute any of this actions. This is a way to try and see what the controller is doing without making any modifications to the registry. The controller can be invoked by using the `registry resolve` command:
```shell
registry resolve projects/demo/locations/global/artifacts/test-manifest --dry-run
```
The above command should generate 30 actions:

* 10 `registry compute lint ...`
* 10 `registry compute vocabulary ...`
* 10 `registry compute complexity ...`
* `registry compute lintstats ...` is omitted because the dependencies for that are non-existent at this point.

##### Standalone mode:
In this mode, we let the controller go through one iteration of the comparison of current and desired states. For each iteration, the controller generates and executes a set of actions which takes the registry closer to it's desired state.

###### Iteration 1:
```shell
registry resolve projects/demo/locations/global/artifacts/test-manifest
```
This will generate and execute the same 30 actions mentioned above. You can see the artifacts generated by the controller by doing a list call
`registry list projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/-`

###### Iteration 2:
Note that  we haven't generated the additional 10 lintstats-spectral artifacts yet. 
```shell
registry resolve projects/demo/locations/global/artifacts/test-manifest
```
After the first iteration, the dependencies (`$resource.spec/artifacts/lint-spectral`) are generated, and hence in this iteration, the controller will generate and execute 10 actions.
* 10 `registry compute lintstats ...`
A quick check with a list call can verify things 
`registry list projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/-`

To summarize, in the standalone mode, the controller can bring the registry up-to-date in two iterations.

##### Continuous mode:
In this mode, the controller is running continuously, making sure that it is always checking the state of the registry in each passing iteration. This can be achieved through a GKE cron job.

###### Step 1: Building the docker image: 
Build the base image which includes the registry tool:
```shell
# Build
docker build -f containers/registry-tools/Dockerfile . -t gcr.io/$REGISTRY_PROJECT_IDENTIFIER/registry-tools

# Push
docker push gcr.io/$REGISTRY_PROJECT_IDENTIFIER/registry-tools:latest
```

Build the linter image from the base image which includes the binaries necessary for the `test-manifest` to work.
```
# Build
envsubst < containers/registry-tools/registry-linters/Dockerfile | docker build -t gcr.io/$REGISTRY_PROJECT_IDENTIFIER/registry-linters -f - .

# Push
docker push gcr.io/$REGISTRY_PROJECT_IDENTIFIER/registry-linters:latest
```

###### Step 2: Deploying the job
Deploy the controller job which will use the `registry-linters` image and run the `registry resolve` command, every 5 mins
```shell
export REGISTRY_MANIFEST_ID=projects/demo/locations/global/artifacts/test-manifest

make deploy-controller-job
```

###### Step 3: Verification
* Once the job is deployed, you should see that in every iteration it is reading the manifest and calculating the actions. 
* Since, we already have the artifacts generated from our previous commands, we need to clean up the previously generated artifacts. 
`registry delete projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/-`.
Once the artifacts are cleaned up, we can see the controller in action.
* Use list command to check what artifacts are generated by the controller, the behavior should be the same as the one described in the standalone case.
`registry list projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/-`
* You now have a controller continuously running in your GKE project which will be responsible for generating and maintaining the artifacts defined in the manifest. 


#### Desired state definitions:

##### Recommended manifest:
To get started with the controller framework, we recommend using the following manifest. It defines four artifacts and all of them can be generated using the `registry tool`.  This is the same manifest used in the walkthrough above.
```yaml
id: "test-manifest"
generated_resources:
  - pattern: "apis/-/versions/-/specs/-/artifacts/lint-spectral"
    dependencies:
      - pattern: "$resource.spec"
        filter: "mime_type.contains('openapi')"
    action: "registry compute lint $resource.spec --linter spectral"
  - pattern: "apis/-/versions/-/specs/-/artifacts/lintstats-spectral"
    dependencies:
      - pattern: "$resource.spec/artifacts/lint-spectral"
    action: "registry compute lintstats $resource.spec --linter spectral"
  - pattern: "apis/-/versions/-/specs/-/artifacts/vocabulary"
    dependencies:
      - pattern: "$resource.spec"
    action: "registry compute vocabulary $resource.spec"
  - pattern: "apis/-/versions/-/specs/-/artifacts/complexity"
    dependencies:
      - pattern: "$resource.spec"
    action: "registry compute complexity $resource.spec"
```

##### Projects with protobuf spec definitions:
If your project contains protobuf definitions, use the api-linter.
```yaml
id: "test-manifest"
generated_resources:
  - pattern: "apis/-/versions/-/specs/-/artifacts/lint-aip"
    dependencies:
      - pattern: "$resource.spec"
        filter: "mime_type.contains('protobuf')"
    action: "registry compute lint $resource.spec --linter aip"
  - pattern: "apis/-/versions/-/specs/-/artifacts/lintstats-aip"
    dependencies:
      - pattern: "$resource.spec/artifacts/lint-aip"
    action: "registry compute lintstats $resource.spec --linter aip"
  - pattern: "apis/-/versions/-/specs/-/artifacts/references"
    dependencies:
      - pattern: "$resource.spec"
        filter: "mime_type.contains('protobuf')"
    action: "registry compute references $resource.spec"
  - pattern: "apis/-/versions/-/specs/-/artifacts/vocabulary"
    dependencies:
      - pattern: "$resource.spec"
    action: "registry compute vocabulary $resource.spec"
  - pattern: "apis/-/versions/-/specs/-/artifacts/complexity"
    dependencies:
      - pattern: "$resource.spec"
    action: "registry compute complexity $resource.spec"
```

