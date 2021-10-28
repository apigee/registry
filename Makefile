lite:
	go install ./...

all:
	./tools/GENERATE-RPC.sh
	./tools/GENERATE-GRPC.sh
	./tools/GENERATE-GAPIC.sh
	./tools/GENERATE-APG.sh
	./tools/GENERATE-ENVOY-DESCRIPTORS.sh
	go install ./...

format:
	go install golang.org/x/tools/cmd/goimports@latest
	goimports -w ./..

apg:
	./tools/GENERATE-APG.sh
	go install ./cmd/apg

protos:
	./tools/GENERATE-RPC.sh
	./tools/GENERATE-GRPC.sh
	./tools/GENERATE-GAPIC.sh
	./tools/GENERATE-ENVOY-DESCRIPTORS.sh

test:
	go clean -testcache
	go test ./...

clean:
	rm -rf docs/ third_party/api-common-protos

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


