include tools/PROTOC-VERSION.sh
export $(shell sed 's/=.*//' tools/PROTOC-VERSION.sh)

lite:
	go install ./...

protoc:
ifeq (, $(shell which protoc))
	@echo "Error! protoc not found on path, please install version ${PROTOC_VERSION}."; exit 1
endif

all: protos
	./tools/GENERATE-CLI.sh
	go install ./...

protos: protoc
	./tools/GENERATE-RPC.sh
	./tools/GENERATE-GRPC.sh
	./tools/GENERATE-GAPIC.sh
	./tools/GENERATE-ENVOY-DESCRIPTORS.sh

lintfix:
	golangci-lint run --fix

test:
	go clean -testcache
	go test ./...

clean:
	rm -rf docs/ third_party/api-common-protos


