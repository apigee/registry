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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"

	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
	"github.com/gogo/googleapis/google/rpc"
)

var (
	grpcport = flag.String("grpcport", ":50051", "grpcport")
	conn     *grpc.ClientConn
	hs       *health.Server
)

const (
	address string = ":50051"
)

// healthServer implements the gRPC health check service.
type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

// authorizationServer implements the Envoy authz service.
type authorizationServer struct{}

// Check implements the check operation in the Envoy authz service.
func (a *authorizationServer) Check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {
	log.Println(">>> Authorization called check()")

	b, err := json.MarshalIndent(req.Attributes.Request.Http.Headers, "", "  ")
	if err == nil {
		log.Println("Inbound Headers: " + string(b))
	}
	ct, err := json.MarshalIndent(req.Attributes.ContextExtensions, "", "  ")
	if err == nil {
		log.Println("Context Extensions: " + string(ct))
	}

	authHeader, ok := req.Attributes.Request.Http.Headers["authorization"]
	if !ok {
		// there's no auth header, so the request is uncredentialed.
		return denyUncredentialedRequest(), nil
	}

	user, err := getUser(authHeader)
	if err != nil || user == nil {
		// we can't find a user for the auth header, so the user is unauthenticated.
		return denyUnauthenticatedUser(), nil
	}

	if user.isWriter() || isReadOnlyMethod(req.Attributes.Request.Http.Headers[":path"]) {
		// the user is authorized so we allow the call.
		return allowAuthorizedUser(user.Email), nil
	}

	// we have a user, but they aren't authorized to do this.
	return denyUnauthorizedUser(), nil
}

// isReadOnlyMethod recognizes Get and List operations as immutable.
func isReadOnlyMethod(path string) bool {
	methodName := filepath.Base(path)
	if strings.HasPrefix(methodName, "Get") ||
		strings.HasPrefix(methodName, "List") {
		return true
	}
	return false
}

// GoogleUser holds information a Google user.
type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	PictureURL    string `json:"picture"`
}

// TODO: read this from a yaml file
var writers = []string{"timburks@google.com", "timburks@gmail.com"}

// isWriter returns true if a user is allowed to make mutable operations.
func (g *GoogleUser) isWriter() bool {
	for _, writer := range writers {
		if g.Email == writer {
			return true
		}
	}
	return false
}

// in-memory cache of users
var users map[string]*GoogleUser

func getUser(authHeader string) (*GoogleUser, error) {
	if users == nil {
		users = make(map[string]*GoogleUser)
	}
	// first check the cache
	cachedUser := users[authHeader]
	if cachedUser != nil {
		return cachedUser, nil
	}
	// otherwise, call the Google userinfo API
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v1/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", authHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected response from auth server: %d", resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	user := &GoogleUser{}
	err = json.Unmarshal(b, user)
	if err != nil {
		return nil, err
	}
	users[authHeader] = user
	return user, nil
}

func allowAuthorizedUser(username string) *auth.CheckResponse {
	return &auth.CheckResponse{
		Status: &rpcstatus.Status{
			Code: int32(rpc.OK),
		},
		HttpResponse: &auth.CheckResponse_OkResponse{
			OkResponse: &auth.OkHttpResponse{
				Headers: []*core.HeaderValueOption{
					{
						Header: &core.HeaderValue{
							Key:   "x-authz-user",
							Value: username,
						},
					},
				},
			},
		},
	}
}

func denyUnauthorizedUser() *auth.CheckResponse {
	return &auth.CheckResponse{
		Status: &rpcstatus.Status{
			Code: int32(rpc.PERMISSION_DENIED),
		},
		HttpResponse: &auth.CheckResponse_DeniedResponse{
			DeniedResponse: &auth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{
					Code: envoy_type.StatusCode_Unauthorized,
				},
				Body: "Permission denied",
			},
		},
	}
}

func denyUnauthenticatedUser() *auth.CheckResponse {
	return &auth.CheckResponse{
		Status: &rpcstatus.Status{
			Code: int32(rpc.UNAUTHENTICATED),
		},
		HttpResponse: &auth.CheckResponse_DeniedResponse{
			DeniedResponse: &auth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{
					Code: envoy_type.StatusCode_Unauthorized,
				},
				Body: "Authorization cannot be validated",
			},
		},
	}
}

func denyUncredentialedRequest() *auth.CheckResponse {
	return &auth.CheckResponse{
		Status: &rpcstatus.Status{
			Code: int32(rpc.UNAUTHENTICATED),
		},
		HttpResponse: &auth.CheckResponse_DeniedResponse{
			DeniedResponse: &auth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{
					Code: envoy_type.StatusCode_Unauthorized,
				},
				Body: "Authorization is missing",
			},
		},
	}
}

func main() {
	flag.Parse()

	if *grpcport == "" {
		fmt.Fprintln(os.Stderr, "missing -grpcport flag (:50051)")
		flag.Usage()
		os.Exit(2)
	}

	lis, err := net.Listen("tcp", *grpcport)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{grpc.MaxConcurrentStreams(10)}
	s := grpc.NewServer(opts...)
	auth.RegisterAuthorizationServer(s, &authorizationServer{})
	healthpb.RegisterHealthServer(s, &healthServer{})

	log.Printf("Starting gRPC Server at %s", *grpcport)
	s.Serve(lis)
}
