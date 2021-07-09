package stability

import (
	"fmt"
	"log"

	pb "github.com/apigee/registry/rpc"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/golang/protobuf/proto"
	"github.com/tufin/oasdiff/diff"
	"gopkg.in/yaml.v3"
)
// tuffin/oasdiff options
type DiffOptions struct{
	prefix string
	filter string
	diffFormat string
	excludeExamples bool
	excludeDescription bool
	summary bool
}

func NewDefaultDiffOptions() DiffOptions{
	diffOptions := DiffOptions{}
	diffOptions.excludeExamples = false
	diffOptions.excludeDescription = false
	diffOptions.summary = true
	return diffOptions
}
//Diff 2 specs with set diff options.
func (diffOptions *DiffOptions)GetDiff(base []byte, revision []byte) ([]byte, error){
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	spec1, err := loader.LoadFromData(base)
	if err != nil {
		fmt.Printf("failed to load base spec from %q with %v", base, err)
		return nil, err
	}
	spec2, err := loader.LoadFromData(revision)
	if err != nil {
		fmt.Printf("failed to load revision spec from %q with %v", revision, err)
		return nil, err
	}
	diffReport, err := diff.Get(&diff.Config{
		ExcludeExamples:    diffOptions.excludeExamples,
		ExcludeDescription: diffOptions.excludeDescription,
		PathFilter:         diffOptions.filter,
		PathPrefix:         diffOptions.prefix,
	}, spec1, spec2)
	if err != nil {
		fmt.Printf("diff failed with %v", err)
		return nil, err
	}
	var yaml_bytes = getYAML(diffReport)
	//Convert Yaml Diff to string.
	stringYamlDiff := string(yaml_bytes)
	p := &pb.TextDiff{
		Diff: stringYamlDiff,
		Spec: spec1.Info.Title,
		SpecRevision: spec2.Info.Title,
	}
	data, err := proto.Marshal(p)
	if err != nil {
		log.Fatal("Marshalling error:", err)
		return nil, err
	}
	return data, nil
}
//Convert Interface to yaml
func getYAML(output interface{}) ([]byte) {
	bytes, err := yaml.Marshal(output)
	if err != nil {
		fmt.Printf("failed to marshal result as yaml with %v", err)
		return bytes
	}
	return bytes
}