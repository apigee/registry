package specdiff

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/apigee/registry/rpc"
	pb "github.com/apigee/registry/rpc"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/diff"
)

// Change TODO
type Change struct{
	FieldPath Stack
	ChangeType string
}

// Stack TODO
type Stack []string

func (s Stack) String() string {
	return strings.Join(s, ".")
}

// IsEmpty TODO
func (s Stack) IsEmpty() bool {
	return len(s) == 0
}

// Push TODO
func (s *Stack) Push(str string) {
	*s = append(*s, str)
}

// Pop TODO
func (s *Stack) Pop() (string, bool) {
	if s.IsEmpty() {
		return "", false
	}
	index := len(*s) - 1
	element := (*s)[index]
	*s = (*s)[:index]
	return element, true
}

// GetDiff takes two yaml or json Open API 3 Specs and diffs them.
// TODO: Return *pb.Diff
func GetDiff(base, revision []byte) (*diff.Diff, error) {
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

func addToDiffProto(diffProto *pb.Diff, change *Change) {
	fieldName := change.FieldPath.String()
	switch change.ChangeType {
	case "added":
		diffProto.Added = append(diffProto.Added, fieldName)
	case "deleted":
		diffProto.Deleted = append(diffProto.Deleted, fieldName)
	}
}

//GetChanges creates a protodif report from a diff.Diff struct.
// TODO: unexport this function or combine it with GetDiff.
func GetChanges(diff *diff.Diff) (*pb.Diff, error){
	diffProto := &pb.Diff{
		Added: []string{},
		Deleted: []string{},
		Modification: make(map[string]*pb.DiffValueModification),
	}
	change := &Change{
		FieldPath:  Stack{},
		ChangeType: "",
	}
	diffNode := reflect.ValueOf(diff)
	err := getChanges(diffNode, diffProto, change)
	return diffProto, err
}

func getChanges(value reflect.Value, diffProto *pb.Diff, change *Change) error {
	// Some values might be pointers, so we should dereference before continuing.
	value = reflect.Indirect(value)

	// Invalid values aren't relevant to the diff.
	if !value.IsValid() {
		return nil
	}

	switch value.Kind() {
	case reflect.Map:
		return searchMapType(value, diffProto, change)
	case reflect.Array, reflect.Slice:
		return searchArrayAndSliceType(value, diffProto, change)
	case reflect.Struct:
		return searchStructType(value, diffProto, change)
	case reflect.Float64, reflect.String, reflect.Bool, reflect.Int:
		valueString, _ := handleAtomicType(value)
		change.FieldPath.Push(valueString)
		addToDiffProto(diffProto, change)
		change.FieldPath.Pop()
		return nil
	default:
		return fmt.Errorf("field %q has unknown type %s with value %v", change.FieldPath, value.Type(), value)
	}
}

func searchMapType(mapNode reflect.Value, diffProto *pb.Diff, change *Change) ( error) {
	if mapNode.Kind() != reflect.Map {
		panic("searchMapType called with invalid type")
	}

	for _ , childNodeKey := range mapNode.MapKeys() {
		childNode := mapNode.MapIndex(childNodeKey)
		if childNode.IsZero(){
			continue
		}

		// TODO: Split this into separate atomic check + parse.
		childNodeKeyName, ok := handleAtomicType(childNodeKey)
		if ok {
			change.FieldPath.Push(childNodeKeyName)
			err := getChanges(childNode, diffProto, change)
			if err != nil {
				return err
			}
			change.FieldPath.Pop()
		} else {
			err := getChanges(childNodeKey, diffProto, change)
			if err != nil {
				return err
			}

			err = getChanges(childNode, diffProto, change)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func searchArrayAndSliceType(arrayNode reflect.Value, diffProto *pb.Diff, change *Change) (error){
	if arrayNode.Kind() != reflect.Slice && arrayNode.Kind() != reflect.Array {
		panic("searchArrayAndSliceType called with invalid type")
	}

	for i := 0; i < (arrayNode.Len()); i++ {
		childNode := arrayNode.Index(i)
		if childNode.IsZero() {
			continue
		}
		if childNode.Kind() == reflect.String {
			change.FieldPath.Push(childNode.String())
			addToDiffProto(diffProto, change)
			change.FieldPath.Pop()
			continue
		}
		err := getChanges(childNode, diffProto, change)
		if err != nil {
			return err
		}
	}
	return nil
}
/*
func searchInterfaceType(interfaceNode reflect.Value, diffProto *pb.Diff, change *Change, err error) (*pb.Diff, *Change, error){
	interfaceElement, _ := handleAtomicType(interfaceNode.Elem())
	change.FieldPath.Push(interfaceElement)
	addToDiffProto(diffProto, change, "","")
	_, _ = change.FieldPath.Pop()
	return diffProto, change, err
}
*/
func searchStructType(structNode reflect.Value, diffProto *pb.Diff, change *Change) (error){
	if structNode.Kind() != reflect.Struct {
		panic("searchStructType called with invalid type")
	}

	if vd, ok := structNode.Interface().(diff.ValueDiff); ok {
		handleValueDiffStruct(vd, diffProto, change)
		return nil
	}

	//emptyChildren := true
	for i := 0; i < structNode.NumField(); i++ {
		tag, ok := structNode.Type().Field(i).Tag.Lookup("json")
		if !ok {
			// Fields that don't have a JSON name aren't part of the diff.
			continue
		}

		// Empty fields in the diff are redundant. Skip them.
		childNode := structNode.Field(i)
		if childNode.IsZero(){
			continue
		}
		// emptyChildren = false

		// Struct field tags have the format "fieldname,omitempty". We only want the field name.
		// TODO: Is there a better way to find this field name?
		fieldName := strings.Split(tag, ",")[0]

		err := handleStructField(childNode, fieldName, diffProto, change)
		if err != nil {
			return err
		}
	}
	// if emptyChildren {
	// 	addToDiffProto(diffProto, change, "","")
	// }
	return nil
}

func handleStructField(value reflect.Value, name string, diffProto *rpc.Diff, change *Change) ( error) {
	// Empty fields in the diff are redundant. Skip them.
	if value.IsZero(){
		return  nil
	}

	switch name {
	case "added", "deleted", "modified":
		change.ChangeType = name
	default:
		change.FieldPath.Push(name)
		defer change.FieldPath.Pop()
	}

	err := getChanges(value, diffProto, change)
	if err != nil {
		return err
	}

	return nil
}

func handleValueDiffStruct(vd diff.ValueDiff, diffProto *pb.Diff, change *Change) {
	fromValue, _ := handleAtomicType(reflect.ValueOf(vd.From))
	toValue, _ := handleAtomicType(reflect.ValueOf(vd.To))

	diffProto.Modification[change.FieldPath.String()] = &pb.DiffValueModification{
		From: fromValue,
		To: toValue,
	}
}

func handleAtomicType(node reflect.Value) (string, bool) {
	switch node.Kind() {
	case reflect.Float64:
		return fmt.Sprintf("%f", node.Float()), true
	case reflect.String:
		return node.String(), true
	case reflect.Bool:
		return fmt.Sprintf("%t", node.Bool()), true
	case reflect.Int:
		return fmt.Sprintf("%d", node.Int()), true
	default:
		return "", false
	}
}
