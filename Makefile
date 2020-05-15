all:	
	cd third_party; sh ./SETUP.sh
	./COMPILE-PROTOS.sh
	go install ./...

test:
	go test -v ./tests/...

clean:
	rm -rf cmd/cli gapic rpc third_party/api-common-protos envoy/proto.pb

build:
	if [ "${FLAME_PROJECT_IDENTIFIER}" == "" ]; then echo "Error! FLAME_PROJECT_IDENTIFIER must be set"; exit -1; fi
	gcloud builds submit --tag gcr.io/${FLAME_PROJECT_IDENTIFIER}/flame-backend

deploy:
	if [ "${FLAME_PROJECT_IDENTIFIER}" == "" ]; then echo "Error! FLAME_PROJECT_IDENTIFIER must be set"; exit -1; fi
	gcloud run deploy flame-backend --image gcr.io/${FLAME_PROJECT_IDENTIFIER}/flame-backend --platform managed
