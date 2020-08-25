package gorm

import (
	"context"
	"log"
	"testing"

	"github.com/apigee/registry/server/models"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func TestCRUD(t *testing.T) {
	c := NewClient()
	ctx := context.TODO()
	var project models.Project
	k := "projects/google"
	c.Get(ctx, &k, &project)
	log.Printf("%+v", project)
	c.Close()
}
