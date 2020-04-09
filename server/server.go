// Copyright 2020 Google LLC. All Rights Reserved.

package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	rpc "apigov.dev/flame/rpc"
	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//
// The server type implements a Flame server.
// Entities are stored using the Cloud Datastore API.
// https://cloud.google.com/datastore/
//
type server struct {
	rpc.UnimplementedFlameServer
}

var credentials *google.Credentials
var projectID string

// newDataStoreClient creates a new data storage connection.
func (s *server) newDataStoreClient(ctx context.Context) (*datastore.Client, error) {
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func newRandomID() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}

// RunServer ...
func RunServer(port string) error {
	ctx := context.TODO()
	var err error
	credentials, err = google.FindDefaultCredentials(ctx)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	projectID = credentials.ProjectID
	if projectID == "" {
		projectID = os.Getenv("FLAME_PROJECT_IDENTIFIER")
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
	rpc.RegisterFlameServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}
