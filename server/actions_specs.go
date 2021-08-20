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

package server

import (
	"context"
	"strings"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/dao"
	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/names"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// CreateApiSpec handles the corresponding API request.
func (s *RegistryServer) CreateApiSpec(ctx context.Context, req *rpc.CreateApiSpecRequest) (*rpc.ApiSpec, error) {
	parent, err := names.ParseVersion(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.GetApiSpec() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid api_spec %+v: body must be provided", req.GetApiSpec())
	}

	return s.createSpec(ctx, parent.Spec(req.GetApiSpecId()), req.GetApiSpec())
}

func (s *RegistryServer) createSpec(ctx context.Context, name names.Spec, body *rpc.ApiSpec) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	if _, err := db.GetSpec(ctx, name); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "API spec %q already exists", name)
	} else if !isNotFound(err) {
		return nil, err
	}

	if err := name.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Creation should only succeed when the parent exists.
	if _, err := db.GetVersion(ctx, name.Version()); err != nil {
		return nil, err
	}

	spec, err := models.NewSpec(name, body)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := db.SaveSpecRevision(ctx, spec); err != nil {
		return nil, err
	}

	if err := db.SaveSpecRevisionContents(ctx, spec, body.GetContents()); err != nil {
		return nil, err
	}

	message, err := spec.BasicMessage(name.String(), []string{})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.notify(ctx, rpc.Notification_CREATED, spec.RevisionName())
	return message, nil
}

// DeleteApiSpec handles the corresponding API request.
func (s *RegistryServer) DeleteApiSpec(ctx context.Context, req *rpc.DeleteApiSpecRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	name, err := names.ParseSpec(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Deletion should only succeed on API specs that currently exist.
	if _, err := db.GetSpec(ctx, name); err != nil {
		return nil, err
	}

	if err := db.DeleteSpec(ctx, name); err != nil {
		return nil, err
	}

	s.notify(ctx, rpc.Notification_DELETED, name.String())
	return &empty.Empty{}, nil
}

// GetApiSpec handles the corresponding API request.
func (s *RegistryServer) GetApiSpec(ctx context.Context, req *rpc.GetApiSpecRequest) (*rpc.ApiSpec, error) {
	if name, err := names.ParseSpec(req.GetName()); err == nil {
		return s.getApiSpec(ctx, name)
	} else if name, err := names.ParseSpecRevision(req.GetName()); err == nil {
		return s.getApiSpecRevision(ctx, name)
	}

	return nil, status.Errorf(codes.InvalidArgument, "invalid resource name %q, must be an API spec or revision", req.GetName())
}

func (s *RegistryServer) getApiSpec(ctx context.Context, name names.Spec) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	spec, err := db.GetSpec(ctx, name)
	if err != nil {
		return nil, err
	}

	tags, err := revisionTags(ctx, db, name.Revision(spec.RevisionID))
	if err != nil {
		return nil, err
	}

	message, err := spec.BasicMessage(name.String(), tags)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return message, nil
}

func (s *RegistryServer) getApiSpecRevision(ctx context.Context, name names.SpecRevision) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	revision, err := db.GetSpecRevision(ctx, name)
	if err != nil {
		return nil, err
	}

	tags, err := revisionTags(ctx, db, name)
	if err != nil {
		return nil, err
	}

	message, err := revision.BasicMessage(name.String(), tags)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return message, nil
}

// GetApiSpecContents handles the corresponding API request.
func (s *RegistryServer) GetApiSpecContents(ctx context.Context, req *rpc.GetApiSpecContentsRequest) (*httpbody.HttpBody, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	if !strings.HasSuffix(req.GetName(), "/contents") {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource name %q, must include /contents suffix", req.GetName())
	}

	var specName = strings.TrimSuffix(req.GetName(), "/contents")
	var spec *models.Spec
	var revisionName names.SpecRevision
	if name, err := names.ParseSpec(specName); err == nil {
		if spec, err = db.GetSpec(ctx, name); err != nil {
			return nil, err
		}
		revisionName = name.Revision(spec.RevisionID)
	} else if name, err := names.ParseSpecRevision(specName); err == nil {
		if spec, err = db.GetSpecRevision(ctx, name); err != nil {
			return nil, err
		}
		revisionName = name
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource name %q, must be an API spec or revision", specName)
	}
	blob, err := db.GetSpecRevisionContents(ctx, revisionName)
	if err != nil {
		return nil, err
	}
	if strings.Contains(spec.MimeType, "+gzip") {
		contents, err := models.GUnzippedBytes(blob.Contents)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "failed to unzip contents with gzip MIME type: %s", err)
		}
		return &httpbody.HttpBody{
			ContentType: strings.Replace(spec.MimeType, "+gzip", "", 1),
			Data:        contents,
		}, nil
	}
	return &httpbody.HttpBody{
		ContentType: spec.MimeType,
		Data:        blob.Contents,
	}, nil
}

// ListApiSpecs handles the corresponding API request.
func (s *RegistryServer) ListApiSpecs(ctx context.Context, req *rpc.ListApiSpecsRequest) (*rpc.ListApiSpecsResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	if req.GetPageSize() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid page_size %d: must not be negative", req.GetPageSize())
	} else if req.GetPageSize() > 1000 {
		req.PageSize = 1000
	} else if req.GetPageSize() == 0 {
		req.PageSize = 50
	}

	parent, err := names.ParseVersion(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	listing, err := db.ListSpecs(ctx, parent, dao.PageOptions{
		Size:   req.GetPageSize(),
		Filter: req.GetFilter(),
		Token:  req.GetPageToken(),
	})
	if err != nil {
		return nil, err
	}

	response := &rpc.ListApiSpecsResponse{
		ApiSpecs:      make([]*rpc.ApiSpec, len(listing.Specs)),
		NextPageToken: listing.Token,
	}

	tags, err := db.GetSpecTags(ctx, parent.Spec("-"))
	if err != nil {
		return nil, err
	}

	tagsByRev := tagsByRevision(tags)
	for i, spec := range listing.Specs {
		response.ApiSpecs[i], err = spec.BasicMessage(spec.Name(), tagsByRev[spec.RevisionName()])
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return response, nil
}

// UpdateApiSpec handles the corresponding API request.
func (s *RegistryServer) UpdateApiSpec(ctx context.Context, req *rpc.UpdateApiSpecRequest) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	if req.GetApiSpec() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid api_spec %+v: body must be provided", req.GetApiSpec())
	} else if err := models.ValidateMask(req.GetApiSpec(), req.GetUpdateMask()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid update_mask %v: %s", req.GetUpdateMask(), err)
	}

	name, err := names.ParseSpec(req.ApiSpec.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	spec, err := db.GetSpec(ctx, name)
	if req.GetAllowMissing() && isNotFound(err) {
		return s.createSpec(ctx, name, req.GetApiSpec())
	} else if err != nil {
		return nil, err
	}

	// Apply the update to the spec - possibly changing the revision ID.
	maskExpansion := models.ExpandMask(req.GetApiSpec(), req.GetUpdateMask())
	if err := spec.Update(req.GetApiSpec(), maskExpansion); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Save the updated/current spec. This creates a new revision or updates the previous one.
	if err := db.SaveSpecRevision(ctx, spec); err != nil {
		return nil, err
	}

	// If the spec contents were updated, save a new blob.
	if len(fieldmaskpb.Intersect(maskExpansion, &fieldmaskpb.FieldMask{Paths: []string{"contents"}}).GetPaths()) > 0 {
		if err := db.SaveSpecRevisionContents(ctx, spec, req.ApiSpec.GetContents()); err != nil {
			return nil, err
		}
	}

	tags, err := revisionTags(ctx, db, name.Revision(spec.RevisionID))
	if err != nil {
		return nil, err
	}

	message, err := spec.BasicMessage(name.String(), tags)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.notify(ctx, rpc.Notification_UPDATED, spec.RevisionName())
	return message, nil
}

func revisionTags(ctx context.Context, db dao.DAO, name names.SpecRevision) ([]string, error) {
	allTags, err := db.GetSpecTags(ctx, name.Spec())
	if err != nil {
		return nil, err
	}

	tags := make([]string, 0)
	for _, tag := range allTags {
		if tag.RevisionID == name.RevisionID {
			tags = append(tags, tag.Tag)
		}
	}

	return tags, nil
}

func tagsByRevision(tags []*models.SpecRevisionTag) map[string][]string {
	revTags := make(map[string][]string, len(tags))
	for _, tag := range tags {
		rev := names.SpecRevision{
			ProjectID:  tag.ProjectID,
			ApiID:      tag.ApiID,
			VersionID:  tag.VersionID,
			SpecID:     tag.SpecID,
			RevisionID: tag.RevisionID,
		}.String()

		revTags[rev] = append(revTags[rev], tag.Tag)
	}

	return revTags
}
