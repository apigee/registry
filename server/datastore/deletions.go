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

package datastore

import (
	"context"
	"log"

	"cloud.google.com/go/datastore"
	"github.com/apigee/registry/server/models"
	"google.golang.org/api/iterator"
)

const verbose = false

// DeleteAllMatches deletes all entities matching a specified query.
func (client *Client) DeleteAllMatches(ctx context.Context, q *Query) error {
	it := client.Run(ctx, q.Distinct())
	key, err := it.Next(nil)
	keys := make([]*datastore.Key, 0)
	for err == nil {
		keys = append(keys, key)
		key, err = it.Next(nil)
		if len(keys) == 500 {
			if verbose {
				log.Printf("Deleting %d %s entities", len(keys), keys[0].Kind)
			}
			err = client.client.DeleteMulti(ctx, keys)
			if err != nil {
				return err
			}
			keys = make([]*datastore.Key, 0)
		}
	}
	if err != iterator.Done {
		return err
	}
	if len(keys) > 0 {
		if verbose {
			log.Printf("Deleting %d %s entities", len(keys), keys[0].Kind)
		}
		return client.client.DeleteMulti(ctx, keys)
	}
	return nil
}

// DeleteChildrenOfProject deletes all the children of a project.
func (client *Client) DeleteChildrenOfProject(ctx context.Context, project *models.Project) error {
	entityNames := []string{
		models.LabelEntityName,
		models.PropertyEntityName,
		models.SpecEntityName,
		models.SpecRevisionTagEntityName,
		models.VersionEntityName,
		models.ApiEntityName,
	}
	for _, entityName := range entityNames {
		q := datastore.NewQuery(entityName)
		q = q.KeysOnly()
		q = q.Filter("ProjectID =", project.ProjectID)
		err := client.DeleteAllMatches(ctx, q)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteChildrenOfApi deletes all the children of a api.
func (client *Client) DeleteChildrenOfApi(ctx context.Context, api *models.Api) error {
	for _, entityName := range []string{models.SpecEntityName, models.VersionEntityName} {
		q := datastore.NewQuery(entityName)
		q = q.KeysOnly()
		q = q.Filter("ProjectID =", api.ProjectID)
		q = q.Filter("ApiID =", api.ApiID)
		err := client.DeleteAllMatches(ctx, q)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteChildrenOfVersion deletes all the children of a version.
func (client *Client) DeleteChildrenOfVersion(ctx context.Context, version *models.Version) error {
	for _, entityName := range []string{models.SpecEntityName} {
		q := datastore.NewQuery(entityName)
		q = q.KeysOnly()
		q = q.Filter("ProjectID =", version.ProjectID)
		q = q.Filter("ApiID =", version.ApiID)
		q = q.Filter("VersionID =", version.VersionID)
		err := client.DeleteAllMatches(ctx, q)
		if err != nil {
			return err
		}
	}
	return nil
}
