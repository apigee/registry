package conformance

import (
	"encoding/json"
	"os/exec"
	"path/filepath"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/rpc"
)

type ApiLinter struct {
	Rules map[string][]string
}

// apiLinterRunner is an interface through which the API Linter executes.
type apiLinterRunner interface {
	// Runs the api-linter with a provided spec and configuration path
	Run(specPath string) ([]*rpc.LintProblem, error)
}

// concreteApiLinterRunner implements the apiLinterRunner interface.
type concreteApiLinterRunner struct{}

func NewApiLinter() ApiLinter {
	return ApiLinter{
		Rules: make(map[string][]string),
	}
}

func unpackGoogleApisProtos(rootDir string) error {
	// Curl the entire folder as a zipped archive from Github (much faster than git checkout).
	curlCmd := exec.Command("curl", "-L", "https://github.com/googleapis/googleapis/archive/refs/heads/master.zip", "-O")
	curlCmd.Dir = rootDir
	err := curlCmd.Run()
	if err != nil {
		return err
	}

	// Unzip the contents of the zipped archive.
	unzipCmd := exec.Command("unzip", "-q", "master.zip")
	unzipCmd.Dir = rootDir
	err = unzipCmd.Run()
	if err != nil {
		return err
	}

	// Move up the google/ directory (the one we're interested in) into the cwd.
	mvCmd := exec.Command("mv", "googleapis-master/google", "google")
	mvCmd.Dir = rootDir
	return mvCmd.Run()
}

func unpackApiCommonProtos(rootDir string) error {
	cmd := exec.Command("git", "clone", "https://github.com/googleapis/api-common-protos")
	cmd.Dir = rootDir
	return cmd.Run()
}

func (linter ApiLinter) AddRule(mimeType string, rule string) error {
	// Check if the linter supports the mime type.
	if !linter.SupportsMimeType(mimeType) {
		return createUnsupportedMimeTypeError(mimeType)
	}

	linter.Rules[mimeType] = append(linter.Rules[mimeType], rule)
	return nil
}

func (linter ApiLinter) GetName() string {
	return "api-linter"
}

// SupportsMimeType returns whether a mime type is supported by API Linter. API
// Linter lints based on AIPs (https://google.aip.dev/) which are defined on protos.
func (linter ApiLinter) SupportsMimeType(mimeType string) bool {
	return core.IsProto(mimeType)
}

// LintSpec lints the spec pointed at by a spec path, which has a provided mime type.
// It returns the results as a LintFile object.
func (linter ApiLinter) LintSpec(
	mimeType string,
	specPath string,
) ([]*rpc.LintProblem, error) {
	return linter.LintSpecImpl(mimeType, specPath, &concreteApiLinterRunner{})
}

func (linter ApiLinter) LintSpecImpl(
	mimeType string,
	specPath string,
	runner apiLinterRunner,
) ([]*rpc.LintProblem, error) {
	// Check if the linter supports the mime type
	if !linter.SupportsMimeType(mimeType) {
		return nil, createUnsupportedMimeTypeError(mimeType)
	}

	// Execute the API linter.
	lintProblems, err := runner.Run(specPath)
	if err != nil {
		return nil, err
	}

	// Filter the problems only those that were enabled by the user.
	return linter.filterProblems(mimeType, lintProblems), nil
}

func (linter ApiLinter) filterProblems(
	mimeType string,
	problems []*rpc.LintProblem) []*rpc.LintProblem {
	// Construct a set of all the problems enabled for this mimetype
	// so we have efficient lookup.
	enabledProblems := make(map[string]bool)
	for _, rule := range linter.Rules[mimeType] {
		enabledProblems[rule] = true
	}

	// From a list of LintProblem objects, only return the rules that were
	// enabled by the caller via `addRule`.
	// We can do this in place.
	n := 0
	for i := 0; i < len(problems); i++ {
		if _, exists := enabledProblems[problems[i].GetRuleId()]; exists {
			problems[n] = problems[i]
			n++
		}
	}

	return problems[:n]
}

func (*concreteApiLinterRunner) Run(specPath string) ([]*rpc.LintProblem, error) {
	// API-linter necessitates being ran on specs in the CWD to avoid many import errors,
	// so we change the directory of the command to the directory of the spec.
	specDirectory := filepath.Dir(specPath)
	specName := filepath.Base(specPath)

	data, err := createAndRunApiLinterCommand(specDirectory, specName)
	if err == nil {
		return parseLinterOutput(data)
	}

	// Unpack api-common-protos and try again if failure occurred
	log.Info("API-linter failed due to an import error, unpacking API common protos and retrying.")
	if err = unpackApiCommonProtos(specDirectory); err == nil {
		data, err = createAndRunApiLinterCommand(specDirectory, specName)
		if err == nil {
			return parseLinterOutput(data)
		}
	}

	log.Info("API-linter failed due to an import error, unpacking GoogleAPIs and retrying.")
	if err = unpackGoogleApisProtos(specDirectory); err == nil {
		data, err = createAndRunApiLinterCommand(specDirectory, specName)
		if err == nil {
			return parseLinterOutput(data)
		}
	}

	return nil, err

}

func createAndRunApiLinterCommand(specDirectory, specName string) ([]byte, error) {
	cmd := exec.Command("api-linter",
		specName,
		"-I", "google",
		"-I", "api-common-protos",
		"--output-format", "json",
	)
	cmd.Dir = specDirectory
	return cmd.CombinedOutput()
}

func parseLinterOutput(data []byte) ([]*rpc.LintProblem, error) {
	// Parse the API Linter output.
	if len(data) == 0 {
		return []*rpc.LintProblem{}, nil
	}
	var lintFiles []*rpc.LintFile
	err := json.Unmarshal(data, &lintFiles)
	if err != nil {
		return nil, err
	}

	// We only passed a single spec to the API linter. Thus
	// the LintFile array should only contain 1 element.
	if len(lintFiles) > 0 {
		lintFile := lintFiles[0]
		return lintFile.GetProblems(), nil
	}
	return []*rpc.LintProblem{}, nil
}
