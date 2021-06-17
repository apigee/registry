package specdiff

import (
	"fmt"
	"reflect"
	"strings"

	pb "github.com/apigee/registry/rpc"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/diff"
)

// Change TODO
type Change struct {
	FieldPath  Stack
	ChangeType string
}

// Stack TODO
type Stack []string

func (s *Stack) String() string {
	return strings.Join(*s, ".")
}

// IsEmpty TODO
func (s *Stack) IsEmpty() bool {
	return len(*s) == 0
}

// Push TODO
func (s *Stack) Push(str string) {
	*s = append(*s, str)
	//fmt.Printf("Pushed | %+v  Stack : %+v \n", str, s)
}

// Pop TODO
func (s *Stack) Pop() (string, bool) {
	if s.IsEmpty() {
		//fmt.Printf("Poped Failed \n")
		return "", false
	}
	index := len(*s) - 1
	element := (*s)[index]
	*s = (*s)[:index]
	return element, true
}

// GetDiff takes two yaml or json Open API 3 Specs and diffs them.
func GetDiff(base []byte, revision []byte) (*diff.Diff, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	baseSpec, err := loader.LoadFromData(base)
	if err != nil {
		err := fmt.Errorf("failed to load base spec from %q with %v", base, err)
		return nil, err
	}
	revisionSpec, err := loader.LoadFromData(revision)
	if err != nil {
		err := fmt.Errorf("failed to load revision spec from %q with %v", revision, err)
		return nil, err
	}
	diffReport, err := diff.Get(&diff.Config{
		ExcludeExamples:    false,
		ExcludeDescription: false,
	}, baseSpec, revisionSpec)
	if err != nil {
		err := fmt.Errorf("diff failed with %v", err)
		return nil, err
	}
	return diffReport, nil
}

// AddToDiffProto TODO
func AddToDiffProto(diffProto *pb.Diff, change *Change, fromValue string, toValue string) {
	if change.ChangeType == "added" {
		diffProto.Added = append(diffProto.Added, change.FieldPath.String())
	}
	if change.ChangeType == "deleted" {
		diffProto.Deleted = append(diffProto.Deleted, change.FieldPath.String())
	}
	if change.ChangeType == "modified" {
		valueModification := &pb.DiffValueModification{
			To:   toValue,
			From: fromValue,
		}
		diffProto.Modification[change.FieldPath.String()] = valueModification
	}
}

//GetChanges creates a protodif report from a diff.Diff struct.
func GetChanges(diff *diff.Diff) (*pb.Diff, error) {
	diffProto := &pb.Diff{
		Added:        []string{},
		Deleted:      []string{},
		Modification: make(map[string]*pb.DiffValueModification),
	}
	change := Change{
		FieldPath:  []string{},
		ChangeType: "",
	}
	diffNode := reflect.ValueOf(diff)
	finaldiffProto, _, err := getChanges(&diffNode, diffProto, &change, nil)
	return finaldiffProto, err
}

func getChanges(value *reflect.Value, diffProto *pb.Diff, change *Change, err error) (*pb.Diff, *Change, error) {
	if err != nil {
		return diffProto, change, err
	}
	//Get the actual Object.
	diffNode := reflect.Indirect(*value)
	if !diffNode.IsValid() {
		return diffProto, change, err
	}
	nodeKind := diffNode.Kind()
	if nodeKind == reflect.Map {
		return searchMapType(diffNode, diffProto, change, err)
	}
	if nodeKind == reflect.Array || nodeKind == reflect.Slice {
		return searchArrayAndSliceType(diffNode, diffProto, change, err)
	}
	/*if nodeKind == reflect.Interface{
		return searchInterfaceType(diffNode, diffProto, change, err)
	}*/
	if nodeKind == reflect.Struct {
		return searchStructType(diffNode, diffProto, change, err)
	}
	atomicType, stringElement := handleNonRecursiveType(diffNode)
	if !atomicType {
		return diffProto, change, fmt.Errorf("NO TYPE:%+v: for Type: %+v\n ", diffNode.Type(), diffNode)
	}

	change.FieldPath.Push(stringElement)
	AddToDiffProto(diffProto, change, "", "")
	change.FieldPath.Pop()

	return diffProto, change, err
}

func searchMapType(mapNode reflect.Value, diffProto *pb.Diff, change *Change, err error) (*pb.Diff, *Change, error) {
	mapNodeKeys := mapNode.MapKeys()
	for _, childNodeKey := range mapNodeKeys {
		childNode := mapNode.MapIndex(childNodeKey)
		if childNode.IsZero() || mapNode.IsNil() {
			continue
		}
		atomicType, childNodeKeyName := handleNonRecursiveType(childNodeKey)
		if atomicType {
			change.FieldPath.Push(childNodeKeyName)
			diffProto, change, err = getChanges(&childNode, diffProto, change, err)
			change.FieldPath.Pop()

		} else {
			diffProto, change, err = getChanges(&childNodeKey, diffProto, change, err)
			diffProto, change, err = getChanges(&childNode, diffProto, change, err)
		}

	}
	return diffProto, change, err
}

func searchArrayAndSliceType(arrayNode reflect.Value, diffProto *pb.Diff, change *Change, err error) (*pb.Diff, *Change, error) {
	for i := 0; i < (arrayNode.Len()); i++ {
		childNode := arrayNode.Index(i)
		if !childNode.IsZero() {
			if childNode.Kind() == reflect.String {
				change.FieldPath.Push(string(childNode.String()))
				AddToDiffProto(diffProto, change, "", "")
				change.FieldPath.Pop()
				continue
			}
			diffProto, change, err = getChanges(&childNode, diffProto, change, err)
		}
	}
	return diffProto, change, err
}

/*func searchInterfaceType(interfaceNode reflect.Value, diffProto *pb.Diff, change *Change, err error)(*pb.Diff, *Change, error){
	_, interfaceElement := handleNonRecursiveType(interfaceNode.Elem())
	change.FieldPath.Push(interfaceElement)
	AddToDiffProto(diffProto, change, "","")
	_, _ = change.FieldPath.Pop()
	return diffProto, change, err
}*/

func searchStructType(structNode reflect.Value, diffProto *pb.Diff, change *Change, err error) (*pb.Diff, *Change, error) {
	if structNode.Type().String() == "diff.ValueDiff" {
		diffProto, change, err = handleValueDiffStruct(structNode, diffProto, change, err)
		return diffProto, change, err
	}
	emptyChildren := true
	for i := 0; i < structNode.NumField(); i++ {
		structNodeTag := structNode.Type().Field(i).Tag

		structNodeTagName, ok := structNodeTag.Lookup("json")
		// Why isnt this an error?
		if !ok {
			continue
		}
		childNode := structNode.Field(i)
		if childNode.IsZero() {
			continue
		}
		structNodeTagName = strings.Split(string(structNodeTagName), ",")[0]
		emptyChildren = false
		setChangeType(structNodeTagName, change)
		if structNodeTagName != "modified" {
			change.FieldPath.Push(structNodeTagName)
			diffProto, change, err = getChanges(&childNode, diffProto, change, err)
			change.FieldPath.Pop()
			continue
		}
		diffProto, change, err = getChanges(&childNode, diffProto, change, err)
	}
	if emptyChildren {
		AddToDiffProto(diffProto, change, "", "")
	}
	return diffProto, change, err
}

func handleValueDiffStruct(valueDiffNode reflect.Value, diffProto *pb.Diff, change *Change, err error) (*pb.Diff, *Change, error) {
	setChangeType("modified", change)
	_, fromValue := handleNonRecursiveType(valueDiffNode.FieldByName("From").Elem())
	_, toValue := handleNonRecursiveType(valueDiffNode.FieldByName("To").Elem())
	AddToDiffProto(diffProto, change, fromValue, toValue)
	return diffProto, change, err
}

func setChangeType(element string, change *Change) {
	if element == "added" {
		change.ChangeType = "added"
	}
	if element == "deleted" {
		change.ChangeType = "deleted"
	}
	if element == "modified" || element == "from" || element == "to" {
		change.ChangeType = "modified"
	}
}

func handleNonRecursiveType(node reflect.Value) (bool, string) {
	nodeKind := node.Kind()
	if nodeKind == reflect.Float64 {
		float64Value := fmt.Sprintf("%f", node.Float())
		return true, string(float64Value)
	}
	if nodeKind == reflect.String {
		return true, node.String()
	}
	if nodeKind == reflect.Bool {
		boolStringValue := fmt.Sprintf("%v", node.Bool())
		return true, boolStringValue
	}
	if nodeKind == reflect.Int {
		intStringValue := fmt.Sprintf("%v", node.Int())
		return true, intStringValue
	}
	return false, ""
}
