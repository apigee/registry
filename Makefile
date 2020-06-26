all:	
	cd third_party; sh ./SETUP.sh
	./COMPILE-PROTOS.sh
	go install ./...

test:
	go clean -testcache
	# test everything except the gapic generated code
	go test `go list ./... | grep -v gapic`

clean:
	rm -rf cmd/cli gapic rpc third_party/api-common-protos envoy/proto.pb

build:
	if [ "${REGISTRY_PROJECT_IDENTIFIER}" == "" ]; then echo "Error! REGISTRY_PROJECT_IDENTIFIER must be set"; exit -1; fi
	gcloud builds submit --tag gcr.io/${REGISTRY_PROJECT_IDENTIFIER}/registry-backend

deploy:
	if [ "${REGISTRY_PROJECT_IDENTIFIER}" == "" ]; then echo "Error! REGISTRY_PROJECT_IDENTIFIER must be set"; exit -1; fi
	gcloud run deploy registry-backend --image gcr.io/${REGISTRY_PROJECT_IDENTIFIER}/registry-backend --platform managed
