// Copyright 2020 Google LLC. All Rights Reserved.

package server

import (
	"context"
	"fmt"
	"net"
	"os"

	rpc "apigov.dev/registry/rpc"
	"cloud.google.com/go/datastore"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// RegistryServer implements a Registry server.
// Entities are stored using the Cloud Datastore API.
// https://cloud.google.com/datastore/
type RegistryServer struct {
	rpc.UnimplementedRegistryServer

	projectID string
}

// newDataStoreClient creates a new data storage connection.
func (s *RegistryServer) newDataStoreClient(ctx context.Context) (*datastore.Client, error) {
	client, err := datastore.NewClient(ctx, s.projectID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// RunServer ...
func RunServer(port string) error {
	ctx := context.TODO()
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	projectID := credentials.ProjectID
	if projectID == "" {
		projectID = os.Getenv("REGISTRY_PROJECT_IDENTIFIER")
	}
	if projectID == "" {
		return fmt.Errorf("unable to determine project ID")
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	reflection.Register(s)
	fmt.Printf("\nServer listening on port %v \n", port)
	rpc.RegisterRegistryServer(s, &RegistryServer{projectID: projectID})
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}
