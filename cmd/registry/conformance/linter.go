package conformance

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
)


type ruleMetadata struct {
	guidelineRule *rpc.Rule //Rule object associated with the linter-rule
	guideline *rpc.Guideline //Guideline object associated with the linter-rule
}

type linterMetadata struct {
	name string
	rules []string
	rulesMetadata map[string]*ruleMetadata
}

func getLinterBinaryName(linterName string) string {
	return "registry-lint-" + linterName
}

func GenerateLinterMetadata(styleguide *rpc.StyleGuide) (map[string]*linterMetadata, error) {

	linterNameToMetadata := make(map[string]*linterMetadata)

	// Iterate through all the guidelines of the style guide.
	for _, guideline := range styleguide.GetGuidelines() {

		// Iterate through all the rules of the style guide.
		for _, rule := range guideline.GetRules() {
			// Get the name of the linter associated with the rule.
			linterName := rule.GetLinter()

			metadata, ok := linterNameToMetadata[linterName]
			if !ok {
				metadata = &linterMetadata{
					name: linterName,
					rules: make([]string, 0),
					rulesMetadata: make(map[string]*ruleMetadata),
				}
				linterNameToMetadata[linterName] = metadata
			}

			//Populate required metadata
			metadata.rules = append(metadata.rules, rule.GetLinterRulename())
			
			if _, ok := metadata.rulesMetadata[rule.GetLinterRulename()]; !ok {
				metadata.rulesMetadata[rule.GetLinterRulename()] = &ruleMetadata{}
			}
			metadata.rulesMetadata[rule.GetLinterRulename()].guideline = guideline
			metadata.rulesMetadata[rule.GetLinterRulename()].guidelineRule = rule
		}
	}

	if len(linterNameToMetadata) == 0 {
		return nil, fmt.Errorf("Empty linter metadata")
	}  
	return linterNameToMetadata, nil
}

func invokeLinter(
	ctx context.Context,
	specDirectory string,
	metadata *linterMetadata) (*rpc.LinterResponse, error){
	
	// Formulate the request.
	requestBytes, err := proto.Marshal(&rpc.LinterRequest{
		SpecDirectory: specDirectory,
		RuleIds:       metadata.rules,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed marshaling linterRequest, Error: %s ", err)
	}

	executableName := getLinterBinaryName(metadata.name)
	cmd := exec.Command(executableName)
	cmd.Stdin = bytes.NewReader(requestBytes)
	cmd.Stderr = os.Stderr
	

	pluginStartTime := time.Now()
	// Run the linter.
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Running the plugin %s return error: %s", executableName, err)
	}

	pluginElapsedTime := time.Since(pluginStartTime)
	log.Debugf(ctx, "Plugin %s ran in time %s", executableName, pluginElapsedTime)

	// Unmarshal the output bytes into a response object. If there's a failure, log and continue.
	linterResponse := &rpc.LinterResponse{}
	err = proto.Unmarshal(output, linterResponse)
	if err != nil {
		return nil, fmt.Errorf("Failed unmarshaling LinterResponse (plugins must write log messages to stderr, not stdout): %s", err)
	}

	// Check if there were any errors in the plugin.
	pluginErrors := make([]string, 0)
	if len(linterResponse.GetErrors()) > 0 {
		for _, err := range linterResponse.GetErrors() {
			pluginErrors = append(pluginErrors, err)
		}
		return nil, fmt.Errorf("Plugin %s encountered errors: %v", executableName, pluginErrors)
	}

	return linterResponse, nil
}