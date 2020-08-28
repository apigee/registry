package gorm

import (
	"context"

	"github.com/apigee/registry/server/models"
)

func (c *Client) DeleteAllMatches(ctx context.Context, q *Query) error {
	return nil
}

// DeleteChildrenOfProject deletes all the children of a project.
func (c *Client) DeleteChildrenOfProject(ctx context.Context, project *models.Project) error {
	return nil
}

// DeleteChildrenOfApi deletes all the children of a api.
func (c *Client) DeleteChildrenOfApi(ctx context.Context, api *models.Api) error {
	return nil
}

// DeleteChildrenOfVersion deletes all the children of a version.
func (c *Client) DeleteChildrenOfVersion(ctx context.Context, version *models.Version) error {
	return nil
}
