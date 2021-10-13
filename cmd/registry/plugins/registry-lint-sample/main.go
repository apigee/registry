package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
)

// Environment contains the environment of a plugin call.
type Environment struct {
	Request    *rpc.LinterRequest  // plugin request object
	Response   *rpc.LinterResponse // response message
	Invocation string              // string representation of call TODO use this
	Verbose    bool                // if true, plugin should log details to stderr
}

// RespondWithErrorAndExit takes in a sequence of errors, sets them in the response,
// responds, and then exits.
func (env *Environment) RespondWithErrorAndExit(errs ...error) {
	errorMessages := make([]string, len(errs))
	for i, err := range errs {
		errorMessages[i] = err.Error()
	}
	env.Response.Errors = append(env.Response.Errors, errorMessages...)
	env.RespondAndExit()
}

// RespondAndExit serializes and writes the plugin response to STDOUT, and then exits.
func (env *Environment) RespondAndExit() {
	responseBytes, _ := proto.Marshal(env.Response)
	os.Stdout.Write(responseBytes)
	os.Exit(0)
}

// NewEnvironment creates a plugin context from arguments and standard input.
func NewEnvironment() (*Environment, error) {
	verbose := flag.Bool("verbose", false, "Write details to stderr.")
	flag.Parse()

	env := &Environment{
		Invocation: os.Args[0],
		Response:   &rpc.LinterResponse{},
		Verbose:    *verbose,
	}

	pluginData, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	if len(pluginData) == 0 {
		if err != nil {
			return nil, fmt.Errorf("no input data")
		}
	}

	// Deserialize the request from the input.
	linterRequest := &rpc.LinterRequest{}
	err = proto.Unmarshal(pluginData, linterRequest)
	if err != nil {
		return nil, err
	}

	// Set the Request in the environment.
	env.Request = linterRequest

	return env, nil
}

func main() {
	env, err := NewEnvironment()
	if err != nil {
		env.RespondWithErrorAndExit(err)
	}

	// Get the API Spec path from the request.
	specPath := env.Request.SpecPath

	// Get the rules to be enabled from the request.
	rules := env.Request.RuleIds

	// Formulate the response. In this sample plugin, we will simply return a fake rule violation /
	// lint problem for every rule that the user specifies, on the given file that is provided.
	lintFile := &rpc.LintFile{
		FilePath: specPath,
	}
	for _, rule := range rules {
		problem := &rpc.LintProblem{}
		problem.RuleId = rule
		problem.Message = fmt.Sprintf("This is a sample violation of the rule %s", rule)
		lintFile.Problems = append(lintFile.Problems, problem)
	}
	lint := &rpc.Lint{
		Name: "registry-lint-sample",
		Files: []*rpc.LintFile{
			lintFile,
		},
	}

	// Set the lint result in the response.
	env.Response.Lint = lint

	// Respond by writing response to STDOUT and exiting.
	env.RespondAndExit()
}
