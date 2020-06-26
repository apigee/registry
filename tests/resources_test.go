package main

import (
	"testing"

	"apigov.dev/registry/models"
)

func TestProductResourceNames(t *testing.T) {
	// Verify that valid names are accepted.
	for _, name := range []string{
		"projects/123/products/abc",
		"projects/1-2_3/products/abc",
	} {
		p, err := models.NewProductFromResourceName(name)
		if err != nil {
			t.Errorf("'%s' is a valid product name but is considered invalid.", name)
		}
		if p != nil {
			resourceName := p.ResourceName()
			if resourceName != name {
				t.Errorf("'%s' failed to round-trip: new name was %s.",
					name, resourceName)
			}
		}
	}
	// verify that invalid names are rejected.
	for _, name := range []string{
		"invalid",
		"projects//products/123",
		"projects/123/products/",
		"projects/123/invalid/123",
		"projects/123/products/ 123",
	} {
		_, err := models.NewProductFromResourceName(name)
		if err == nil {
			t.Errorf("'%s' is an invalid product name but is considered valid.", name)
		}
	}
}

func TestVersionResourceNames(t *testing.T) {
	// Verify that valid names are accepted.
	for _, name := range []string{
		"projects/123/products/abc/versions/123",
		"projects/1-2_3/products/abc/versions/123",
	} {
		p, err := models.NewVersionFromResourceName(name)
		if err != nil {
			t.Errorf("'%s' is a valid version name but is considered invalid.", name)
		}
		if p != nil {
			resourceName := p.ResourceName()
			if resourceName != name {
				t.Errorf("'%s' failed to round-trip: new name was %s.",
					name, resourceName)
			}
		}
	}
	// verify that invalid names are rejected.
	for _, name := range []string{
		"invalid",
		"projects//products/123",
		"projects/123/products/",
		"projects/123/invalid/123",
		"projects/123/products/ 123",
	} {
		_, err := models.NewVersionFromResourceName(name)
		if err == nil {
			t.Errorf("'%s' is an invalid version name but is considered valid.", name)
		}
	}
}

func TestSpecResourceNames(t *testing.T) {
	// Verify that valid names are accepted.
	for _, name := range []string{
		"projects/123/products/abc/versions/123/specs/abc",
		"projects/1-2_3/products/abc/versions/123/specs/abc",
	} {
		p, err := models.NewSpecFromResourceName(name)
		if err != nil {
			t.Errorf("'%s' is a valid spec name but is considered invalid.", name)
		}
		if p != nil {
			resourceName := p.ResourceName()
			if resourceName != name {
				t.Errorf("'%s' failed to round-trip: new name was %s.",
					name, resourceName)
			}
		}
	}
	// verify that invalid names are rejected.
	for _, name := range []string{
		"invalid",
		"projects//products/123",
		"projects/123/products/",
		"projects/123/invalid/123",
		"projects/123/products/ 123",
	} {
		_, err := models.NewSpecFromResourceName(name)
		if err == nil {
			t.Errorf("'%s' is an invalid spec name but is considered valid.", name)
		}
	}
}
