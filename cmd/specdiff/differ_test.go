package specdiff

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/tufin/oasdiff/diff"
	"google.golang.org/protobuf/testing/protocmp"
)

type testStruct struct {
	Name    string            `json:"name,omitempty"`
	TestMap map[string]string `json:"testmap,omitempty"`
}

func TestDiffProtoStruct(t *testing.T) {
	tests := []struct {
		desc         string
		baseSpec     string
		revisionSpec string
		wantProto    *rpc.Diff
	}{
		{
			desc:         "Struct Diff Added",
			baseSpec:     "./test-specs/base-test.yaml",
			revisionSpec: "./test-specs/struct-test-add.yaml",
			wantProto: &rpc.Diff{
				Added: []string{
					"components.schemas.Pet.required.added.age",
					"components.schemas.Pet.properties.added.age",
				},
				Deleted:      []string{},
				Modification: map[string]*rpc.DiffValueModification{},
			},
		},
		{
			desc:         "Struct Diff Deleted",
			baseSpec:     "./test-specs/base-test.yaml",
			revisionSpec: "./test-specs/struct-test-delete.yaml",
			wantProto: &rpc.Diff{
				Added: []string{},
				Deleted: []string{
					"components.schemas.Pet.required.deleted.name",
				},
				Modification: map[string]*rpc.DiffValueModification{},
			},
		},
		{
			desc:         "Struct Diff Modified",
			baseSpec:     "./test-specs/base-test.yaml",
			revisionSpec: "./test-specs/struct-test-modify.yaml",
			wantProto: &rpc.Diff{
				Added:   []string{},
				Deleted: []string{},
				Modification: map[string]*rpc.DiffValueModification{
					"info.version": {
						To:   "1.0.1",
						From: "1.0.0",
					},
					"components.schemas.Pet.properties.tag.type": {
						To:   "integer",
						From: "string",
					},
					"components.schemas.Pet.properties.tag.format": {
						To: "int64",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			yamlFile, err := ioutil.ReadFile(test.baseSpec)
			if err != nil {
				t.Fatalf("Failed to get test-yaml: %s", err)
			}
			yamlFile2, err := ioutil.ReadFile(test.revisionSpec)
			if err != nil {
				t.Fatalf("Failed to get test-yaml: %s", err)
			}
			diff, err := GetDiff(yamlFile, yamlFile2)
			if err != nil {
				t.Fatalf("Failed to get get diff.Diff: %s", err)
			}
			diffProto, err := GetChanges(diff)
			if err != nil {
				t.Fatalf("Failed to get get diff proto: %s", err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
			}
			if !cmp.Equal(test.wantProto, diffProto, opts) {
				t.Errorf("Test %+v returned unexpected diff (-want +got):\n%s", test.desc, cmp.Diff(test.wantProto, diffProto, opts))
			}
		})
	}
}

func TestMaps(t *testing.T) {
	tests := []struct {
		desc      string
		testMap   reflect.Value
		change    Change
		wantProto *rpc.Diff
	}{
		{
			desc: "Map Diff Simple",
			testMap: reflect.ValueOf(map[string]string{
				"input1": "result1",
				"input2": "result2",
			}),
			change: Change{
				FieldPath:  []string{},
				ChangeType: "added",
			},
			wantProto: &rpc.Diff{
				Added: []string{
					"input1.result1",
					"input2.result2",
				},
			},
		}, {
			desc: "Map Diff Recursive Key",
			testMap: reflect.ValueOf(map[*testStruct]testStruct{
				{
					Name: "TestStruct1",
					TestMap: map[string]string{
						"input1": "result1",
						"input2": "result2",
					},
				}: {
					Name: "TestStructResult1",
				},
			}),
			change: Change{
				FieldPath:  []string{},
				ChangeType: "added",
			},
			wantProto: &rpc.Diff{
				Added: []string{
					"testmap.input1.result1",
					"testmap.input2.result2",
					"name.TestStruct1",
					"name.TestStructResult1",
				},
			},
		},
	}

	opts := cmp.Options{protocmp.Transform(), cmpopts.SortSlices(func(a, b string) bool { return a < b })}
	for _, test := range tests {
		val := test.testMap
		diffProto := &rpc.Diff{
			Added:        []string{},
			Deleted:      []string{},
			Modification: make(map[string]*rpc.DiffValueModification),
		}
		change := test.change
		diffProtoResult, _, err := searchMapType(val, diffProto, &change, nil)
		if err != nil {
			t.Fatalf("Failed to get get diff proto: %+v", err)
		}
		if !cmp.Equal(test.wantProto, diffProtoResult, opts) {
			t.Errorf("Test %+v returned unexpected diff (-want +got):\n%s", test.desc, cmp.Diff(test.wantProto, diffProto, opts))
		}
	}
}

func TestArrays(t *testing.T) {
	tests := []struct {
		desc      string
		testArray reflect.Value
		change    Change
		wantProto *rpc.Diff
	}{
		{
			desc: "Array Diff Test Simple",
			testArray: reflect.ValueOf([]string{
				"input1",
				"input2",
				"input3",
				"input4",
			}),
			change: Change{
				FieldPath:  []string{},
				ChangeType: "added",
			},
			wantProto: &rpc.Diff{
				Added: []string{
					"input1",
					"input2",
					"input3",
					"input4",
				},
			},
		},
	}

	opts := cmp.Options{protocmp.Transform(), cmpopts.SortSlices(func(a, b string) bool { return a < b })}
	for _, test := range tests {
		val := test.testArray
		diffProto := &rpc.Diff{
			Added:        []string{},
			Deleted:      []string{},
			Modification: make(map[string]*rpc.DiffValueModification),
		}
		change := test.change
		diffProtoResult, _, err := searchArrayAndSliceType(val, diffProto, &change, nil)
		if err != nil {
			t.Fatalf("Failed to get get diff proto: %+v", err)
		}
		if !cmp.Equal(test.wantProto, diffProtoResult, opts) {
			t.Errorf("Test %+v returned unexpected diff (-want +got):\n%s", test.desc, cmp.Diff(test.wantProto, diffProto, opts))
		}
	}
}

func TestValueDiff(t *testing.T) {
	tests := []struct {
		desc          string
		testValueDiff reflect.Value
		change        Change
		wantProto     *rpc.Diff
	}{
		{
			desc: "Value Diff Test",
			testValueDiff: reflect.ValueOf(diff.ValueDiff{From: 66,
				To: true}),
			change: Change{
				FieldPath:  []string{"ValueDiffTest"},
				ChangeType: "Modified",
			},
			wantProto: &rpc.Diff{
				Modification: map[string]*rpc.DiffValueModification{
					"ValueDiffTest": {
						To:   "true",
						From: "66"},
				},
			},
		},
	}

	opts := cmp.Options{protocmp.Transform(), cmpopts.SortSlices(func(a, b string) bool { return a < b })}
	for _, test := range tests {
		val := test.testValueDiff
		diffProto := &rpc.Diff{
			Added:        []string{},
			Deleted:      []string{},
			Modification: make(map[string]*rpc.DiffValueModification),
		}
		change := test.change
		diffProtoResult, _, err := handleValueDiffStruct(val, diffProto, &change, nil)
		if err != nil {
			t.Fatalf("Failed to get get diff proto: %+v", err)
		}
		if !cmp.Equal(test.wantProto, diffProtoResult, opts) {
			t.Errorf("Test %+v returned unexpected diff (-want +got):\n%s", test.desc, cmp.Diff(test.wantProto, diffProto, opts))
		}
	}
}

//Test Yaml Diff
/*func TestDiffYaml(t *testing.T){
	yamlFile, err := ioutil.ReadFile("./test-specs/base-test.yaml")
	if err != nil {
		t.Fatalf("Failed to get test-yaml: %+v", err)
		t.FailNow()
	}
	yamlFile2, err := ioutil.ReadFile("./test-specs/struct-test-modify.yaml")
	if err != nil {
		t.Logf("Failed to get test-yaml: %+v", err)
		t.FailNow()
	}
	diff, _ := GetDiff(yamlFile, yamlFile2)

	diffProto, err := GetChangesRecursive(diff)
	for i := 0; i < len(diffProto.Added); i++{
		t.Logf("CHANGETYPE:%+v 	 |Change: %+v \n", "Added", diffProto.Added[i])
	}
	for i := 0; i < len(diffProto.Deleted); i++{
		t.Logf("CHANGETYPE:%+v 	 |Change: %+v \n", "Deleted", diffProto.Deleted[i])
	}
	for i := range diffProto.Modification{
		t.Logf("CHANGETYPE:%+v 	 |Change: %+v |Modification: %+v \n", "Modified", i, diffProto.Modification[i])
	}
	//'t.Logf("Protobuf: %+v \n", diffProto)

	t.FailNow()
}

func getYAML(output interface{}) ([]byte) {
	bytes, err := yaml.Marshal(output)
	if err != nil {
		fmt.Printf("failed to marshal result as yaml with %v", err)
		return bytes
	}
	return bytes
}
*/
