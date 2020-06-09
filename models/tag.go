// Copyright 2020 Google LLC. All Rights Reserved.

package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	rpc "apigov.dev/registry/rpc"
	ptypes "github.com/golang/protobuf/ptypes"
)

// TagEntityName is used to represent tags in the datastore.
const TagEntityName = "Tag"

// TagsRegexp returns a regular expression that matches collection of tags.
func TagsRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/tags$")
}

// TagRegexp returns a regular expression that matches a Tag resource name.
func TagRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/tags/" + nameRegex + "$")
}

// Tag ...
type Tag struct {
	ProjectID  string    // Project associated with tag (required).
	ProductID  string    // Product associated with tag (if appropriate).
	VersionID  string    // Version associated with tag (if appropriate).
	SpecID     string    // Spec associated with tag (if appropriate).
	TagID      string    // Tag identifier (required).
	CreateTime time.Time // Creation time.
	UpdateTime time.Time // Time of last change.
	Subject    string    // Subject of the tag.
}

// NewTagFromParentAndTagID returns an initialized tag for a specified parent and tagID.
func NewTagFromParentAndTagID(parent string, tagID string) (*Tag, error) {
	// Return an error if the tagID is invalid.
	if err := validateID(tagID); err != nil {
		return nil, err
	}
	// Match regular expressions to identify the parent of this tag.
	var m [][]string
	// Is the parent a project?
	m = ProjectRegexp().FindAllStringSubmatch(parent, -1)
	if m != nil {
		return &Tag{
			ProjectID: m[0][1],
			TagID:     tagID,
			Subject:   parent,
		}, nil
	}
	// Is the parent a product?
	m = ProductRegexp().FindAllStringSubmatch(parent, -1)
	if m != nil {
		return &Tag{
			ProjectID: m[0][1],
			ProductID: m[0][2],
			TagID:     tagID,
			Subject:   parent,
		}, nil
	}
	// Is the parent a version?
	m = VersionRegexp().FindAllStringSubmatch(parent, -1)
	if m != nil {
		return &Tag{
			ProjectID: m[0][1],
			ProductID: m[0][2],
			VersionID: m[0][3],
			TagID:     tagID,
			Subject:   parent,
		}, nil
	}
	// Is the parent a spec?
	m = SpecRegexp().FindAllStringSubmatch(parent, -1)
	if m != nil {
		return &Tag{
			ProjectID: m[0][1],
			ProductID: m[0][2],
			VersionID: m[0][3],
			SpecID:    m[0][4],
			TagID:     tagID,
			Subject:   parent,
		}, nil
	}
	// Return an error for an unrecognized parent.
	return nil, fmt.Errorf("invalid parent '%s'", parent)
}

// NewTagFromResourceName parses resource names and returns an initialized tag.
func NewTagFromResourceName(name string) (*Tag, error) {
	// split name into parts
	parts := strings.Split(name, "/")
	if parts[len(parts)-2] != "tags" {
		return nil, fmt.Errorf("invalid tag name '%s'", name)
	}
	// build tag from parent and tagID
	parent := strings.Join(parts[0:len(parts)-2], "/")
	tagID := parts[len(parts)-1]
	return NewTagFromParentAndTagID(parent, tagID)
}

// ResourceName generates the resource name of a tag.
func (tag *Tag) ResourceName() string {
	switch {
	case tag.SpecID != "":
		return fmt.Sprintf("projects/%s/products/%s/versions/%s/specs/%s/tags/%s",
			tag.ProjectID, tag.ProductID, tag.VersionID, tag.SpecID, tag.TagID)
	case tag.VersionID != "":
		return fmt.Sprintf("projects/%s/products/%s/versions/tags/%s",
			tag.ProjectID, tag.ProductID, tag.VersionID, tag.TagID)
	case tag.ProductID != "":
		return fmt.Sprintf("projects/%s/products/%s/tags/%s",
			tag.ProjectID, tag.ProductID, tag.TagID)
	case tag.ProjectID != "":
		return fmt.Sprintf("projects/%s/tags/%s",
			tag.ProjectID, tag.TagID)
	default:
		return "UNKNOWN"
	}
}

// Message returns a message representing a tag.
func (tag *Tag) Message() (message *rpc.Tag, err error) {
	message = &rpc.Tag{}
	message.Name = tag.ResourceName()
	message.Subject = tag.Subject
	message.CreateTime, err = ptypes.TimestampProto(tag.CreateTime)
	message.UpdateTime, err = ptypes.TimestampProto(tag.UpdateTime)
	return message, err
}

// Update modifies a tag using the contents of a message.
func (tag *Tag) Update(message *rpc.Tag) error {
	tag.Subject = message.GetSubject()
	tag.UpdateTime = tag.CreateTime
	return nil
}
