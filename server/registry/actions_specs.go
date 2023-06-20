// Copyright 2020 Google LLC.
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

package registry

import (
	"context"
	"strings"

	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/internal/storage"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// CreateApiSpec handles the corresponding API request.
func (s *RegistryServer) CreateApiSpec(ctx context.Context, req *rpc.CreateApiSpecRequest) (*rpc.ApiSpec, error) {
	// Parent name must be valid.
	parent, err := names.ParseVersion(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// Spec name must be valid.
	name := parent.Spec(req.GetApiSpecId())
	if err := name.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var response *rpc.ApiSpec
	if err := s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		var err error
		response, err = s.createSpec(ctx, db, name, req.GetApiSpec())
		return err
	}); err != nil {
		return nil, err
	}
	s.notify(ctx, rpc.Notification_CREATED, response.GetName())
	return response, nil
}

func (s *RegistryServer) createSpec(ctx context.Context, db *storage.Client, name names.Spec, body *rpc.ApiSpec) (*rpc.ApiSpec, error) {
	// The spec must not already exist.
	if _, err := db.GetSpec(ctx, name); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "API spec %q already exists", name)
	} else if !isNotFound(err) {
		return nil, err
	}

	spec, err := models.NewSpec(name, body)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := db.CreateSpecRevision(ctx, spec); err != nil {
		return nil, err
	}

	if err := db.SaveSpecRevisionContents(ctx, spec, body.GetContents()); err != nil {
		return nil, err
	}

	return spec.BasicMessage(name.String())
}

// DeleteApiSpec handles the corresponding API request.
func (s *RegistryServer) DeleteApiSpec(ctx context.Context, req *rpc.DeleteApiSpecRequest) (*emptypb.Empty, error) {
	// Spec name must be valid.
	name, err := names.ParseSpec(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		return db.LockSpecs(ctx).DeleteSpec(ctx, name, req.GetForce())
	}); err != nil {
		return nil, err
	}
	s.notify(ctx, rpc.Notification_DELETED, req.GetName())
	return &emptypb.Empty{}, nil
}

// GetApiSpec handles the corresponding API request.
func (s *RegistryServer) GetApiSpec(ctx context.Context, req *rpc.GetApiSpecRequest) (*rpc.ApiSpec, error) {
	s.begin()
	defer s.end()
	if name, err := names.ParseSpec(req.GetName()); err == nil {
		return s.getApiSpec(ctx, name)
	} else if name, err := names.ParseSpecRevision(req.GetName()); err == nil {
		return s.getApiSpecRevision(ctx, name)
	}

	return nil, status.Errorf(codes.InvalidArgument, "invalid resource name %q, must be an API spec or revision", req.GetName())
}

func (s *RegistryServer) getApiSpec(ctx context.Context, name names.Spec) (*rpc.ApiSpec, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	spec, err := db.GetSpec(ctx, name)
	if err != nil {
		return nil, err
	}

	message, err := spec.BasicMessage(name.String())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return message, nil
}

func (s *RegistryServer) getApiSpecRevision(ctx context.Context, name names.SpecRevision) (*rpc.ApiSpec, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	revision, err := db.GetSpecRevision(ctx, name)
	if err != nil {
		return nil, err
	}

	message, err := revision.BasicMessage(name.String())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return message, nil
}

// GetApiSpecContents handles the corresponding API request.
func (s *RegistryServer) GetApiSpecContents(ctx context.Context, req *rpc.GetApiSpecContentsRequest) (*httpbody.HttpBody, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	var specName = req.GetName()
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

	if strings.Contains(spec.MimeType, "+gzip") && !incomingContextAllowsGZIP(ctx) {
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
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

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

	listing, err := db.ListSpecs(ctx, parent, storage.PageOptions{
		Size:   req.GetPageSize(),
		Filter: req.GetFilter(),
		Order:  req.GetOrderBy(),
		Token:  req.GetPageToken(),
	})
	if err != nil {
		return nil, err
	}

	response := &rpc.ListApiSpecsResponse{
		ApiSpecs:      make([]*rpc.ApiSpec, len(listing.Specs)),
		NextPageToken: listing.Token,
	}

	for i, spec := range listing.Specs {
		response.ApiSpecs[i], err = spec.BasicMessage(spec.Name())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return response, nil
}

// UpdateApiSpec handles the corresponding API request.
func (s *RegistryServer) UpdateApiSpec(ctx context.Context, req *rpc.UpdateApiSpecRequest) (*rpc.ApiSpec, error) {
	// Spec body must be valid.
	if req.GetApiSpec() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid api_spec %+v: body must be provided", req.GetApiSpec())
	}
	// Spec name must be valid.
	name, err := names.ParseSpec(req.ApiSpec.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// Update mask must be valid.
	if err := models.ValidateMask(req.GetApiSpec(), req.GetUpdateMask()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid update_mask %v: %s", req.GetUpdateMask(), err)
	}
	var response *rpc.ApiSpec
	if err = s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		spec, err := db.GetSpec(ctx, name)
		if err == nil {
			// Apply the update to the spec - possibly changing the revision ID.
			maskExpansion := models.ExpandMask(req.GetApiSpec(), req.GetUpdateMask())
			if err := spec.Update(req.GetApiSpec(), maskExpansion); err != nil {
				return err
			}
			// Save the updated/current spec. This creates a new revision or updates the previous one.
			if err := db.SaveSpecRevision(ctx, spec); err != nil {
				return err
			}
			// If the spec contents were updated, save a new blob.
			if len(fieldmaskpb.Intersect(maskExpansion, &fieldmaskpb.FieldMask{Paths: []string{"contents"}}).GetPaths()) > 0 {
				if err := db.SaveSpecRevisionContents(ctx, spec, req.ApiSpec.GetContents()); err != nil {
					return err
				}
			}
			response, err = spec.BasicMessage(name.String())
			return err
		} else if status.Code(err) == codes.NotFound && req.GetAllowMissing() {
			response, err = s.createSpec(ctx, db, name, req.GetApiSpec())
			if status.Code(err) == codes.AlreadyExists {
				err = status.Error(codes.Aborted, err.Error())
			}
			return err
		} else {
			return err
		}
	}); err != nil {
		return nil, err
	}
	s.notify(ctx, rpc.Notification_UPDATED, response.GetName())
	return response, nil
}

func incomingContextAllowsGZIP(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	for _, a := range md["accept-encoding"] {
		if a == "gzip" {
			return true
		}
	}
	return false
}
