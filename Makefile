lite:
	go install ./...

all:	protos
	go install ./...

protos:
	cd third_party; sh ./SETUP.sh
	./tools/COMPILE-PROTOS.sh

test:
	go clean -testcache
	go test -v ./...

clean:
	rm -rf \
		cmd/apg/*.go gapic/*.go rpc/*.go \
		third_party/api-common-protos third_party/gnostic \
		envoy/proto.pb

build:
ifndef REGISTRY_PROJECT_IDENTIFIER
	@echo "Error! REGISTRY_PROJECT_IDENTIFIER must be set."; exit 1
endif
ifndef DB_CONFIG
	$(eval DB_CONFIG := registry) # default to use config/registry.yaml
endif
	gcloud builds submit . --substitutions _REGISTRY_PROJECT_IDENTIFIER="${REGISTRY_PROJECT_IDENTIFIER}",_DB_CONFIG=$(DB_CONFIG)

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

index:
	gcloud datastore indexes create index.yaml

index-cleanup:
	gcloud datastore indexes cleanup index.yaml

build-workers:
ifndef REGISTRY_PROJECT_IDENTIFIER
	@echo "Error! REGISTRY_PROJECT_IDENTIFIER must be set."; exit 1
endif
	gcloud builds submit --config deployments/capabilities/cloudbuild.yaml \
    --substitutions _REGISTRY_PROJECT_IDENTIFIER="${REGISTRY_PROJECT_IDENTIFIER}"

deploy-workers:
	./deployments/capabilities/DEPLOY-WORKERS.sh