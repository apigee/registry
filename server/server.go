// Copyright 2020 Google Inc. All Rights Reserved.

package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	rpc "apigov.dev/flame/rpc"
	"cloud.google.com/go/datastore"
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

// newDataStoreClient creates a new data storage connection.
func (s *server) newDataStoreClient(ctx context.Context) (*datastore.Client, error) {
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return nil, err
	}
	projectID := credentials.ProjectID
	if projectID == "" {
		return nil, errors.New("unable to determine project ID")
	}
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// RunServer ...
func RunServer(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	reflection.Register(s)
	fmt.Printf("\nServer listening on port %v \n", port)
	rpc.RegisterFlameServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
