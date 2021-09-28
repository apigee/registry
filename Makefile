lite:
	go install ./...

all:
	./tools/GENERATE-RPC.sh
	./tools/GENERATE-GAPIC.sh
	./tools/GENERATE-APG.sh
	./tools/GENERATE-ENVOY-DESCRIPTORS.sh
	go install ./...

apg:
	./tools/GENERATE-APG.sh

docs:
	./tools/GENERATE-DOCS.sh

envoy-descriptors:
	./tools/GENERATE-ENVOY-DESCRIPTORS.sh

lint:
	./tools/LINT-PROTOS.sh

openapi:
	./tools/GENERATE-OPENAPI.sh

protos:
	./tools/GENERATE-RPC.sh
	./tools/GENERATE-GAPIC.sh

test:
	go clean -testcache
	go test ./...

clean:
	rm -rf cmd/apg/*.go gapic/*.go rpc/*.go third_party/api-common-protos envoy/proto.pb

build:
ifndef REGISTRY_PROJECT_IDENTIFIER
	@echo "Error! REGISTRY_PROJECT_IDENTIFIER must be set."; exit 1
endif
	gcloud builds submit . --substitutions _REGISTRY_PROJECT_IDENTIFIER="${REGISTRY_PROJECT_IDENTIFIER}"

deploy:
ifndef REGISTRY_PROJECT_IDENTIFIER
	@echo "Error! REGISTRY_PROJECT_IDENTIFIER must be set."; exit 1
endif
	gcloud run deploy registry-backend --image gcr.io/${REGISTRY_PROJECT_IDENTIFIER}/registry-backend --platform managed

deploy-gke:
ifndef REGISTRY_PROJECT_IDENTIFIER
	@echo "Error! REGISTRY_PROJECT_IDENTIFIER must be set."; exit 1
endif
ifeq ($(LB),internal)
	./deployments/gke/DEPLOY-TO-GKE.sh deployments/gke/service-internal.yaml
else
	./deployments/gke/DEPLOY-TO-GKE.sh
endif

build-workers:
ifndef REGISTRY_PROJECT_IDENTIFIER
	@echo "Error! REGISTRY_PROJECT_IDENTIFIER must be set."; exit 1
endif
	gcloud builds submit --config deployments/capabilities/cloudbuild.yaml \
    --substitutions _REGISTRY_PROJECT_IDENTIFIER="${REGISTRY_PROJECT_IDENTIFIER}"

deploy-workers:
	./deployments/capabilities/DEPLOY-WORKERS.sh

# Actions for controller
deploy-controller-job:
ifndef REGISTRY_PROJECT_IDENTIFIER
	@echo "Error! REGISTRY_PROJECT_IDENTIFIER must be set."; exit 1
endif
ifndef REGISTRY_MANIFEST_ID
	@echo "Error! REGISTRY_MANIFEST_ID must be set."; exit 1
endif
	gcloud container clusters get-credentials registry-backend --zone us-central1-a
	envsubst < deployments/controller/gke-job/cron-job.yaml | kubectl apply -f -

deploy-controller-dashboard:
ifndef REGISTRY_PROJECT_IDENTIFIER
	@echo "Error! REGISTRY_PROJECT_IDENTIFIER must be set."; exit 1
endif
	./deployments/controller/dashboard/DEPLOY.sh

deploy-controller: deploy-controller-job deploy-controller-dashboard


