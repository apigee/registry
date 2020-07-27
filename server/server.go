// Copyright 2020 Google LLC. All Rights Reserved.

package server

import (
	"context"
	"fmt"
	"net"
	"os"

	"apigov.dev/registry/rpc"
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

	projectID     string
	storageClient *datastore.Client
}

// getDataStoreClient returns a shared data storage connection.
func (s *RegistryServer) getDataStoreClient(ctx context.Context) (*datastore.Client, error) {
	if s.storageClient != nil {
		return s.storageClient, nil
	}
	client, err := datastore.NewClient(ctx, s.projectID)
	if err != nil {
		return nil, err
	}
	s.storageClient = client
	return client, nil
}

// if we had one client per handler, this would close the client.
func (r *RegistryServer) releaseDataStoreClient(client *datastore.Client) {
}

func getProjectID() (string, error) {
	ctx := context.TODO()
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}
	projectID := credentials.ProjectID
	if projectID == "" {
		projectID = os.Getenv("REGISTRY_PROJECT_IDENTIFIER")
	}
	if projectID == "" {
		return "", fmt.Errorf("unable to determine project ID")
	}
	return projectID, nil
}

// RunServer runs the Registry server on a specified port
func RunServer(port string) error {
	// Get project ID to use in registry server.
	projectID, err := getProjectID()
	if err != nil {
		return err
	}
	// Construct registry server.
	server := grpc.NewServer()
	reflection.Register(server)
	rpc.RegisterRegistryServer(server, &RegistryServer{projectID: projectID})
	// Create a listener and use it to run the server.
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	fmt.Printf("\nServer listening on port %v \n", port)
	return server.Serve(listener)
}
