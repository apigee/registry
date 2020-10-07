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
	ParentName         string
	ResourceName       string
	ViewEnumName       string
	ExtraRequestFields string
	HasRevisions       bool
}

// Service is a top-level description of a CRUD API service.
type Service struct {
	Service  string
	Version  string
	Entities []Entity
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
		"collection_path": func(parentName, pluralEntityName string) string {
			if parentName == "" {
				return "/" + version + "/" + strings.ToLower(pluralEntityName)
			}
			return "/" + version + "/{parent=" + parentName + "}/" + strings.ToLower(pluralEntityName)
		},
		"resource_path": func(resourceName string) string {
			return "/" + version + "/{name=" + resourceName + "}"
		},
		"resource_path_for_update": func(entityName, resourceName string) string {
			return "/" + version + "/{" + strings.ToLower(entityName) + ".name=" + resourceName + "}"
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
				ParentName:         "",
				ResourceName:       "projects/*",
				ViewEnumName:       "",
				ExtraRequestFields: "",
				HasRevisions:       false,
			},
			{
				Name:               "Api",
				PluralName:         "Apis",
				ParentName:         "projects/*",
				ResourceName:       "projects/*/apis/*",
				ViewEnumName:       "",
				ExtraRequestFields: "",
				HasRevisions:       false,
			},
			{
				Name:               "Version",
				PluralName:         "Versions",
				ParentName:         "projects/*/apis/*",
				ResourceName:       "projects/*/apis/*/versions/*",
				ViewEnumName:       "",
				ExtraRequestFields: "",
				HasRevisions:       false,
			},
			{
				Name:               "Spec",
				PluralName:         "Specs",
				ParentName:         "projects/*/apis/*/versions/*",
				ResourceName:       "projects/*/apis/*/versions/*/specs/*",
				ViewEnumName:       "View",
				ExtraRequestFields: "",
				HasRevisions:       true,
			},
			{
				Name:               "Property",
				PluralName:         "Properties",
				ParentName:         "projects/*",
				ResourceName:       "projects/*/properties/*",
				ViewEnumName:       "View",
				ExtraRequestFields: "",
				HasRevisions:       false,
			},
			{
				Name:               "Label",
				PluralName:         "Labels",
				ParentName:         "projects/*",
				ResourceName:       "projects/*/labels/*",
				ViewEnumName:       "",
				ExtraRequestFields: "",
				HasRevisions:       false,
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
