// Copyright 2020 Google LLC. All Rights Reserved.

package models

import (
	"context"
	"fmt"
	"regexp"
	"time"

	rpc "apigov.dev/registry/rpc"
	"cloud.google.com/go/datastore"
	ptypes "github.com/golang/protobuf/ptypes"
)

// ProductEntityName is used to represent products in the datastore.
const ProductEntityName = "Product"

// ProductsRegexp returns a regular expression that matches collection of products.
func ProductsRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/products$")
}

// ProductRegexp returns a regular expression that matches a product resource name.
func ProductRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/products/" + nameRegex + "$")
}

// Product ...
type Product struct {
	ProjectID          string    // Uniquely identifies a project.
	ProductID          string    // Uniquely identifies a product within a project.
	DisplayName        string    // A human-friendly name.
	Description        string    // A detailed description.
	CreateTime         time.Time // Creation time.
	UpdateTime         time.Time // Time of last change.
	Availability       string    // Availability of the API.
	RecommendedVersion string    // Recommended API version.
}

// ParseParentProject ...
func ParseParentProject(parent string) ([]string, error) {
	r := regexp.MustCompile("^projects/" + nameRegex + "$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid project '%s'", parent)
	}
	return m[0], nil
}

// NewProductFromParentAndProductID returns an initialized product for a specified parent and productID.
func NewProductFromParentAndProductID(parent string, productID string) (*Product, error) {
	r := regexp.MustCompile("^projects/" + nameRegex + "$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid parent '%s'", parent)
	}
	if err := validateID(productID); err != nil {
		return nil, err
	}
	product := &Product{}
	product.ProjectID = m[0][1]
	product.ProductID = productID
	return product, nil
}

// NewProductFromResourceName parses resource names and returns an initialized product.
func NewProductFromResourceName(name string) (*Product, error) {
	product := &Product{}
	m := ProductRegexp().FindAllStringSubmatch(name, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid product name (%s)", name)
	}
	product.ProjectID = m[0][1]
	product.ProductID = m[0][2]
	return product, nil
}

// ResourceName generates the resource name of a product.
func (product *Product) ResourceName() string {
	return fmt.Sprintf("projects/%s/products/%s", product.ProjectID, product.ProductID)
}

// Message returns a message representing a product.
func (product *Product) Message() (message *rpc.Product, err error) {
	message = &rpc.Product{}
	message.Name = product.ResourceName()
	message.DisplayName = product.DisplayName
	message.Description = product.Description
	message.CreateTime, err = ptypes.TimestampProto(product.CreateTime)
	message.UpdateTime, err = ptypes.TimestampProto(product.UpdateTime)
	message.Availability = product.Availability
	message.RecommendedVersion = product.RecommendedVersion
	return message, err
}

// Update modifies a product using the contents of a message.
func (product *Product) Update(message *rpc.Product) error {
	product.DisplayName = message.GetDisplayName()
	product.Description = message.GetDescription()
	product.Availability = message.GetAvailability()
	product.RecommendedVersion = message.GetRecommendedVersion()
	product.UpdateTime = time.Now()
	return nil
}

// DeleteChildren deletes all the children of a product.
func (product *Product) DeleteChildren(ctx context.Context, client *datastore.Client) error {
	for _, entityName := range []string{SpecEntityName, VersionEntityName} {
		q := datastore.NewQuery(entityName)
		q = q.KeysOnly()
		q = q.Filter("ProjectID =", product.ProjectID)
		q = q.Filter("ProductID =", product.ProductID)
		err := deleteAllMatches(ctx, client, q)
		if err != nil {
			return err
		}
	}
	return nil
}
