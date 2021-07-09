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
					"components.schemas.Pet.required.age",
					"components.schemas.Pet.properties.age",
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
					"components.schemas.Pet.required.name",
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
				t.Fatalf("Failed to get base spec yaml: %s", err)
			}
			yamlFile2, err := ioutil.ReadFile(test.revisionSpec)
			if err != nil {
				t.Fatalf("Failed to get revision spec yaml: %s", err)
			}
			diffProto, err := GetDiff(yamlFile, yamlFile2)
			if err != nil {
				t.Fatalf("Failed to get diff.Diff: %s", err)
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
		change    change
		wantProto *rpc.Diff
	}{
		{
			desc: "Map Diff String",
			testMap: reflect.ValueOf(map[string]string{
				"input1": "result1",
				"input2": "result2",
			}),
			change: change{
				fieldPath:  []string{},
				changeType: "added",
			},
			wantProto: &rpc.Diff{
				Added: []string{
					"input1.result1",
					"input2.result2",
				},
			},
		}, {
			desc: "Map Diff Recursive Key",
			testMap: reflect.ValueOf(map[diff.Endpoint]testStruct{
				{ Method : "Test-Method",
					Path: "Test-Path",
				}: {
					Name: "TestStructResult1",
				},
			}),
			change: change{
				fieldPath:  []string{},
				changeType: "added",
			},
			wantProto: &rpc.Diff{
				Added: []string{
					"{method.Test-Method path.Test-Path}.name.TestStructResult1",
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
		err := searchMapType(val, diffProto, &change)
		if err != nil {
			t.Fatalf("Failed to get diff proto, returnd with error: %+v", err)
		}
		if !cmp.Equal(test.wantProto, diffProto, opts) {
			t.Errorf("Test %+v returned unexpected diff (-want +got):\n%s", test.desc, cmp.Diff(test.wantProto, diffProto, opts))
		}
	}
}

func TestArrays(t *testing.T) {
	tests := []struct {
		desc      string
		testArray reflect.Value
		change    change
		wantProto *rpc.Diff
	}{
		{
			desc: "Array Diff String",
			testArray: reflect.ValueOf([]string{
				"input1",
				"input2",
				"input3",
				"input4",
				"",
			}),
			change: change{
				fieldPath:  []string{},
				changeType: "added",
			},
			wantProto: &rpc.Diff{
				Added: []string{
					"input1",
					"input2",
					"input3",
					"input4",
				},
			},
		},{
			desc: "Array Diff Endpoint",
			testArray: reflect.ValueOf([]diff.Endpoint{
				{
				Method : "Test-Method-1",
				Path: "Test-Path-1",
				},{
				Method : "Test-Method-2",
				Path: "Test-Path-2",
				},{
				Method : "Test-Method-3",
				Path: "Test-Path-3",
				},
			}),
			change: change{
				fieldPath:  []string{},
				changeType: "added",
			},
			wantProto: &rpc.Diff{
				Added: []string{
					"{method.Test-Method-1 path.Test-Path-1}",
					"{method.Test-Method-2 path.Test-Path-2}",
					"{method.Test-Method-3 path.Test-Path-3}",
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
		err := searchArrayAndSliceType(val, diffProto, &change)
		if err != nil {
			t.Fatalf("Failed to get get diff proto: %+v", err)
		}
		if !cmp.Equal(test.wantProto, diffProto, opts) {
			t.Errorf("Test %+v returned unexpected diff (-want +got):\n%s", test.desc, cmp.Diff(test.wantProto, diffProto, opts))
		}
	}
}

func TestValueDiff(t *testing.T) {
	tests := []struct {
		desc          string
		testValueDiff reflect.Value
		change        change
		wantProto     *rpc.Diff
	}{
		{
			desc: "Value Diff Test",
			testValueDiff: reflect.ValueOf(diff.ValueDiff{From: 66,
				To: true}),
			change: change{
				fieldPath:  []string{"ValueDiffTest"},
				changeType: "Modified",
			},
			wantProto: &rpc.Diff{
				Modification: map[string]*rpc.DiffValueModification{
					"ValueDiffTest": {
						To:   "true",
						From: "66",
					},
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
		vd, ok := val.Interface().(diff.ValueDiff)
		if !ok {
			t.Fatalf("Failed to convert from reflect.Value to diff.ValueDiff")
		}
		handleValueDiffStruct(vd, diffProto, &change)
		if !cmp.Equal(test.wantProto, diffProto, opts) {
			t.Errorf("Test %+v returned unexpected diff (-want +got):\n%s", test.desc, cmp.Diff(test.wantProto, diffProto, opts))
		}
	}
}

//Test Yaml Diff
/*
func TestDiffYaml(t *testing.T){
	yamlFile, err := ioutil.ReadFile("./data/openapi-test1.yaml")
	if err != nil {
		t.Fatalf("Failed to get test-yaml: %+v", err)
	}
	yamlFile2, err := ioutil.ReadFile("./data/openapi-test2.yaml")
	if err != nil {
		t.Fatalf("Failed to get test-yaml: %+v", err)
	}

	diff, _ := GetDiff(yamlFile, yamlFile2)

	diffProto, err := GetChanges(diff)
	for i := 0; i < len(diffProto.Added); i++{
		t.Logf("CHANGETYPE:%+v 	 |Change: %+v \n", "Added", diffProto.Added[i])
	}
	for i := 0; i < len(diffProto.Deleted); i++{
		t.Logf("CHANGETYPE:%+v 	 |Change: %+v \n", "Deleted", diffProto.Deleted[i])
	}
	for i := range diffProto.Modification{
		t.Logf("CHANGETYPE:%+v 	 |Change: %+v |Modification: %+v \n", "Modified", i, diffProto.Modification[i])
	}


	t.FailNow()
}

func getYAML(output interface{}) (string) {
	bytes, err := yaml.Marshal(output)
	if err != nil {
		fmt.Printf("failed to marshal result as yaml with %v", err)
		return string(bytes)
	}
	return string(bytes)
}
*/
