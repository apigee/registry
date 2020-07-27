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
				Name:               "Product",
				PluralName:         "Products",
				ParentName:         "projects/*",
				ResourceName:       "projects/*/products/*",
				ViewEnumName:       "",
				ExtraRequestFields: "",
				HasRevisions:       false,
			},
			{
				Name:               "Version",
				PluralName:         "Versions",
				ParentName:         "projects/*/products/*",
				ResourceName:       "projects/*/products/*/versions/*",
				ViewEnumName:       "",
				ExtraRequestFields: "",
				HasRevisions:       false,
			},
			{
				Name:               "Spec",
				PluralName:         "Specs",
				ParentName:         "projects/*/products/*/versions/*",
				ResourceName:       "projects/*/products/*/versions/*/specs/*",
				ViewEnumName:       "SpecView",
				ExtraRequestFields: "",
				HasRevisions:       true,
			},
			{
				Name:               "Property",
				PluralName:         "Properties",
				ParentName:         "projects/*",
				ResourceName:       "projects/*/properties/*",
				ExtraRequestFields: "",
				HasRevisions:       false,
			},
			{
				Name:               "Label",
				PluralName:         "Labels",
				ParentName:         "projects/*",
				ResourceName:       "projects/*/labels/*",
				ExtraRequestFields: "",
				HasRevisions:       false,
			},
		},
	}

	f, err := os.Create("../" + filename)
	check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	err = t.ExecuteTemplate(w, filename, service)
	if err != nil {
		log.Printf("%+v", err)
	}
	w.Flush()

}
