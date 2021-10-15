package lint

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
)

// RespondWithErrorAndExit takes in a sequence of errors, sets them in the response,
// responds, and then exits.
func RespondWithErrorAndExit(errs ...error) {
	errorMessages := make([]string, len(errs))
	for i, err := range errs {
		errorMessages[i] = err.Error()
	}
	response := &rpc.LinterResponse{
		Errors: errorMessages,
	}
	RespondAndExit(response)
}

// RespondAndExit serializes and writes the plugin response to STDOUT, and then exits.
func RespondAndExit(response *rpc.LinterResponse) {
	responseBytes, _ := proto.Marshal(response)
	os.Stdout.Write(responseBytes)
	os.Exit(0)
}

// GetRequest constructs a LinterRequest object from standard input.
func GetRequest() (*rpc.LinterRequest, error) {
	// Read from stdin.
	pluginData, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	if len(pluginData) == 0 && err != nil {
		return nil, fmt.Errorf("no input data")
	}

	// Deserialize the request from the input into a request object.
	linterRequest := &rpc.LinterRequest{}
	if err := proto.Unmarshal(pluginData, linterRequest); err != nil {
		return nil, err
	}

	return linterRequest, nil
}

func Main(runner LinterPluginRunner) {
	req, err := GetRequest()
	if err != nil {
		RespondWithErrorAndExit(err)
	}

	resp, err := runner.Run(req)
	if err != nil {
		RespondWithErrorAndExit(err)
	}

	RespondAndExit(resp)
}
