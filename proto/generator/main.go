// Generates flame_service.proto
package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"text/template"
)

const filename = "flame_service.proto"
const version = "v1alpha1"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Entity is an entity in a CRUD API service.
type Entity struct {
	Name         string
	PluralName   string
	ParentName   string
	ResourceName string
	ViewEnumName string
}

// Service is a top-level description of a CRUD API service.
type Service struct {
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
	}).ParseFiles("flame_service.tmpl")
	check(err)

	service := Service{
		Entities: []Entity{
			{
				Name:         "Product",
				PluralName:   "Products",
				ParentName:   "projects/*",
				ResourceName: "projects/*/products/*",
				ViewEnumName: "",
			},
			{
				Name:         "Version",
				PluralName:   "Versions",
				ParentName:   "projects/*/products/*",
				ResourceName: "projects/*/products/*/versions/*",
				ViewEnumName: "",
			},
			{
				Name:         "Spec",
				PluralName:   "Specs",
				ParentName:   "projects/*/products/*/versions/*",
				ResourceName: "projects/*/products/*/versions/*/specs/*",
				ViewEnumName: "",
			},
			{
				Name:         "File",
				PluralName:   "Files",
				ParentName:   "projects/*/products/*/versions/*/specs/*",
				ResourceName: "projects/*/products/*/versions/*/specs/*/files/*",
				ViewEnumName: "FileView",
			},
			{
				Name:         "Property",
				PluralName:   "Properties",
				ParentName:   "projects/*",
				ResourceName: "projects/*/properties/*",
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
