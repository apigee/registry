all:	
	cd third_party; sh ./SETUP.sh
	./COMPILE-PROTOS.sh
	go install ./...

test:
	go test -v ./tests/...

clean:
	rm -rf cmd/cli gapic rpc third_party/api-common-protos envoy/proto.pb

build:
	gcloud builds submit --tag gcr.io/${FLAME_PROJECT_IDENTIFIER}/flame-backend

deploy:
	gcloud run deploy --image gcr.io/${FLAME_PROJECT_IDENTIFIER}/flame-backend --platform managed
