all:	
	cd third_party; sh ./SETUP.sh
	./COMPILE.sh
	go install ./cmd/...

test:
	go test -v ./tests/...

clean:
	rm -rf cmd/cli gapic rpc third_party/api-common-protos envoy/proto.pb
