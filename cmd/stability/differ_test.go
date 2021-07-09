package stability

import (
	"io/ioutil"
	"testing"

	pb "github.com/apigee/registry/rpc"
	"github.com/golang/protobuf/proto"
)

const (
	expected_diff =
			`info:
    version:
        from: 1.0.0
        to: 1.0.1
components:
    schemas:
        modified:
            Pet:
                required:
                    deleted:
                        - name
                properties:
                    added:
                        - added
                        - age
`
	expected_name_1 = "Test"
	expected_name_2 = "Test"
)
//Test Yaml Diff
func TestDiffYaml(t *testing.T){
	yamlFile, err := ioutil.ReadFile("test1-3.0.0.yaml")
	if err != nil {
		t.Logf("Failed to get test-yaml: %+v", err)
		t.FailNow()
	}
	yamlFile2, err := ioutil.ReadFile("test2-3.0.0.yaml")
	if err != nil {
		t.Logf("Failed to get test-yaml: %+v", err)
		t.FailNow()
	}
	expected_diff_proto := &pb.TextDiff{
		Diff: expected_diff,
		Spec: expected_name_1,
		SpecRevision: expected_name_2,
	}
	diffOptions := NewDefaultDiffOptions()
	seralized_diff_proto, err := diffOptions.GetDiff(yamlFile, yamlFile2)
	if err != nil {
		t.Logf("GetDiff error: %v", err)
		t.FailNow()
	}
	diff_proto := &pb.TextDiff{}
	err = proto.Unmarshal(seralized_diff_proto, diff_proto)
	if err != nil {
		t.Logf("Marshalling error: %v", err)
		t.FailNow()
	}
	if diff_proto.Diff != expected_diff_proto.Diff {
		t.Errorf("The Diffs were not equal |\n" +
			"want: %v \ngot: %v", expected_diff_proto.Diff, diff_proto.Diff)
	}
	if diff_proto.Spec != expected_diff_proto.Spec {
		t.Errorf("The Spec names were not equal |\n" +
				"want: %v \ngot: %v", expected_diff_proto.Spec, diff_proto.Spec)
	}
	if diff_proto.SpecRevision != expected_diff_proto.SpecRevision {
		t.Errorf("The Revision names were not equal |\n" +
				"want: %v \ngot: %v", expected_diff_proto.SpecRevision,
				diff_proto.SpecRevision)
	}
}
//Test Json Diff
func TestDiffJson(t *testing.T){
	jsonFile, err := ioutil.ReadFile("test1-3.0.0.json")
	if err != nil {
		t.Logf("Failed to get test-yaml: %+v", err)
		t.FailNow()
	}
	jsonFile2, err := ioutil.ReadFile("test2-3.0.0.json")
	if err != nil {
		t.Logf("Failed to get test-yaml: %+v", err)
		t.FailNow()
	}
	expected_diff_proto := &pb.TextDiff{
		Diff: expected_diff,
		Spec: expected_name_1,
		SpecRevision: expected_name_2,
	}
	diffOptions := NewDefaultDiffOptions()
	seralized_diff_proto, err := diffOptions.GetDiff(jsonFile, jsonFile2)
	if err != nil {
		t.Logf("GetDiff error: %v", err)
		t.FailNow()
	}
	diff_proto := &pb.TextDiff{}
	err = proto.Unmarshal(seralized_diff_proto, diff_proto)
	if err != nil {
		t.Logf("Marshalling error: %v", err)
		t.FailNow()
	}
	if diff_proto.Diff != expected_diff_proto.Diff {
		t.Errorf("The Diffs were not equal |\n" +
				"want: %v \ngot: %v", expected_diff_proto.Diff, diff_proto.Diff)
	}
	if diff_proto.Spec != expected_diff_proto.Spec {
		t.Errorf("The Spec names were not equal |\n" +
				"want: %v \ngot: %v", expected_diff_proto.Spec, diff_proto.Spec)
	}
	if diff_proto.SpecRevision != expected_diff_proto.SpecRevision {
		t.Errorf("The Revison names were not equal |\n" +
				"want: %v \ngot: %v", expected_diff_proto.SpecRevision,
			diff_proto.SpecRevision)
	}
}

