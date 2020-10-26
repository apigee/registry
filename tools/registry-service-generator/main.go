// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Generates registry_service.proto
package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"text/template"
)

const filename = "registry_service.proto"
const version = "v1alpha1"
const service = "google.cloud.apigee.registry"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Entity is an entity in a CRUD API service.
type Entity struct {
	Name               string
	PluralName         string
	ParentNames        []string
	ChildName          string
	ViewEnumName       string
	ExtraRequestFields string
	HasRevisions       bool
	HasFieldMasks      bool
	HasUpdate          bool
}

// Service is a top-level description of a CRUD API service.
type Service struct {
	Service  string
	Version  string
	Entities []Entity
}

func resourceName(parentName, childName string) string {
	if parentName != "" {
		return parentName + "/" + childName
	}
	return childName
}

func main() {
	t, err := template.New("").Funcs(template.FuncMap{
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"lower_comment": func(s string) string {
			if strings.HasPrefix(s, "Api") {
				return strings.Replace(s, "Api", "API", 1)
			}
			return strings.ToLower(s)
		},
		"first": func(names []string) string {
			if len(names) > 0 {
				return names[0]
			}
			return ""
		},
		"rest": func(names []string) []string {
			if len(names) > 0 {
				return names[1:]
			} else {
				return nil
			}
		},
		"resource_name": resourceName,
		"collection_path": func(parentName, pluralEntityName string) string {
			if parentName == "" {
				return "/" + version + "/" + strings.ToLower(pluralEntityName)
			}
			return "/" + version + "/{parent=" + parentName + "}/" + strings.ToLower(pluralEntityName)
		},
		"resource_path": func(parentName, childName string) string {
			return "/" + version + "/{name=" + resourceName(parentName, childName) + "}"
		},
		"resource_path_for_update": func(entityName, parentName, childName string) string {
			return "/" + version + "/{" + strings.ToLower(entityName) + ".name=" + resourceName(parentName, childName) + "}"
		},
		"path_for_service": func(service string) string {
			return strings.Replace(service, ".", "/", -1)
		},
	}).ParseFiles("registry_service.tmpl")
	check(err)

	service := Service{
		Service: service,
		Version: version,
		Entities: []Entity{
			{
				Name:               "Project",
				PluralName:         "Projects",
				ParentNames:        []string{},
				ChildName:          "projects/*",
				ViewEnumName:       "",
				ExtraRequestFields: "",
				HasRevisions:       false,
				HasFieldMasks:      true,
				HasUpdate:          true,
			},
			{
				Name:               "Api",
				PluralName:         "Apis",
				ParentNames:        []string{"projects/*"},
				ChildName:          "apis/*",
				ViewEnumName:       "",
				ExtraRequestFields: "",
				HasRevisions:       false,
				HasFieldMasks:      true,
				HasUpdate:          true,
			},
			{
				Name:               "Version",
				PluralName:         "Versions",
				ParentNames:        []string{"projects/*/apis/*"},
				ChildName:          "versions/*",
				ViewEnumName:       "",
				ExtraRequestFields: "",
				HasRevisions:       false,
				HasFieldMasks:      true,
				HasUpdate:          true,
			},
			{
				Name:               "Spec",
				PluralName:         "Specs",
				ParentNames:        []string{"projects/*/apis/*/versions/*"},
				ChildName:          "specs/*",
				ViewEnumName:       "View",
				ExtraRequestFields: "",
				HasRevisions:       true,
				HasFieldMasks:      true,
				HasUpdate:          true,
			},
			{
				Name:       "Property",
				PluralName: "Properties",
				ParentNames: []string{
					"projects/*",
					"projects/*/apis/*",
					"projects/*/apis/*/versions/*",
					"projects/*/apis/*/versions/*/specs/*",
				},
				ChildName:          "properties/*",
				ViewEnumName:       "View",
				ExtraRequestFields: "",
				HasRevisions:       false,
				HasFieldMasks:      false,
				HasUpdate:          true,
			},
			{
				Name:       "Label",
				PluralName: "Labels",
				ParentNames: []string{
					"projects/*",
					"projects/*/apis/*",
					"projects/*/apis/*/versions/*",
					"projects/*/apis/*/versions/*/specs/*",
				},
				ChildName:          "labels/*",
				ViewEnumName:       "",
				ExtraRequestFields: "",
				HasRevisions:       false,
				HasFieldMasks:      false,
				HasUpdate:          false,
			},
		},
	}

	f, err := os.Create(filename)
	check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	err = t.ExecuteTemplate(w, filename, service)
	if err != nil {
		log.Printf("%+v", err)
	}
	w.Flush()

}
