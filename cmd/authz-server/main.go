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
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/gogo/googleapis/google/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v2"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	portFlag   = flag.String("p", ":50051", "port")
	configFlag = flag.String("c", "", "configuration file")
)

// AuthzConfig configures the authz filter.
type AuthzConfig struct {
	Anonymous bool     `json:"anonymous" yaml:"anonymous"`
	TrustJWTs bool     `json:"trustJWTs" yaml:"trustJWTs"`
	Readers   []string `json:"readers" yaml:"readers"`
	Writers   []string `json:"writers" yaml:"writers"`
	// hard-coded tokens and corresponding user ids (for testing only)
	Tokens map[string]string `json:"tokens" yaml:"tokens"`
}

var config AuthzConfig

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
	response, err := a.check(ctx, req)
	if err == nil {
		log.Printf(">>> Authorization response %+v", response)
	} else {
		log.Printf(">>> Authorization error %s", err.Error())
	}
	return response, err
}

func (a *authorizationServer) check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {
	b, err := json.MarshalIndent(req.Attributes.Request.Http.Headers, "", "  ")
	if err == nil {
		log.Println("Inbound Headers: " + string(b))
	}

	authHeader, ok := req.Attributes.Request.Http.Headers["authorization"]
	if !ok {
		// there's no auth header, so the request is uncredentialed.
		if config.Anonymous {
			return allowOrDenyUser("anonymous", req)
		}
		return denyUncredentialedRequest(), nil
	}
	re := regexp.MustCompile("^[bB]earer[ ]+(.*)$")
	m := re.FindStringSubmatch(authHeader)
	if m == nil {
		return denyMalformedCredentials(), nil
	}
	credential := m[1]

	userid, ok := config.Tokens[credential]
	if ok {
		return allowOrDenyUser(userid, req)
	}

	isJWT, signature := isJWTToken(credential)
	if isJWT {
		if config.TrustJWTs {
			// get the user email from the token
			email := getJWTTokenEmail(credential)
			if email != "" {
				return allowOrDenyUser(email, req)
			}
		}
		// log a modified signature (this will cause verification to fail)
		if signature == "SIGNATURE_REMOVED_BY_GOOGLE" {
			log.Printf("Token has altered signature: %s", signature)
		}
		// try to verify an identity token
		token, err := getVerifiedToken(credential)
		if err != nil {
			return nil, err
		}
		if err == nil && token != nil {
			return allowOrDenyUser(token.Email, req)
		}
	} else {
		// try to verify an access token
		user, err := getUser(credential)
		if err != nil {
			return nil, err
		}
		if err == nil && user != nil {
			return allowOrDenyUser(user.Email, req)
		}
	}

	// we can't find a user for the auth header, so the user is unauthenticated.
	return denyUnauthenticatedUser(), nil
}

func allowOrDenyUser(email string, req *auth.CheckRequest) (*auth.CheckResponse, error) {
	if isReadOnlyMethod(
		req.Attributes.Request.Http.Headers[":path"],
		req.Attributes.Request.Http.Headers[":method"],
	) {
		if isReader(email) {
			return allowAuthorizedUser(email), nil
		}
	} else {
		if isWriter(email) {
			return allowAuthorizedUser(email), nil
		}
	}
	// we have a user, but they aren't authorized to do this.
	return denyUnauthorizedUser(), nil
}

// isReadOnlyMethod recognizes Get and List operations as immutable.
func isReadOnlyMethod(path, method string) bool {
	// assume that "GET" is for http requests to read-only methods.
	if method == "GET" {
		return true
	}
	// for grpc requests, decide based on the method name.
	methodName := filepath.Base(path)
	if strings.HasPrefix(methodName, "Get") ||
		strings.HasPrefix(methodName, "List") {
		return true
	}
	return false
}

// isReader returns true if a user is allowed to make immutable operations.
func isReader(email string) bool {
	for _, reader := range config.Readers {
		m, err := filepath.Match(reader, email)
		if m && err == nil {
			return true
		}
	}
	return false
}

// isWriter returns true if a user is allowed to make mutable operations.
func isWriter(email string) bool {
	for _, writer := range config.Writers {
		m, err := filepath.Match(writer, email)
		if m && err == nil {
			return true
		}
	}
	return false
}

type jwtTokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func isJWTToken(credential string) (bool, string) {
	parts := strings.Split(credential, ".")
	if len(parts) != 3 {
		return false, ""
	}
	header := parts[0]
	v, err := base64.RawURLEncoding.DecodeString(header)
	if err != nil {
		return false, ""
	}
	var tokenHeader jwtTokenHeader
	_ = json.Unmarshal(v, &tokenHeader)
	valid := tokenHeader.Typ == "JWT"
	signature := parts[2]
	return valid, signature
}

type jwtTokenPayload struct {
	Email string `json:"email"`
}

func getJWTTokenEmail(credential string) string {
	parts := strings.Split(credential, ".")
	if len(parts) != 3 {
		return ""
	}
	payload := parts[1]
	v, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return ""
	}
	log.Printf("token payload %+v\n", string(v))
	var tokenPayload jwtTokenPayload
	_ = json.Unmarshal(v, &tokenPayload)
	return tokenPayload.Email
}

// GoogleUser holds information about a Google user.
type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	PictureURL    string `json:"picture"`
}

// in-memory cache of users
var users map[string]*GoogleUser
var mutex sync.Mutex

func getUser(credential string) (*GoogleUser, error) {
	mutex.Lock()
	if users == nil {
		users = make(map[string]*GoogleUser)
	}
	// first check the cache
	cachedUser := users[credential]
	mutex.Unlock()
	if cachedUser != nil {
		log.Printf("cached user: %+v for %s", cachedUser, credential)
		return cachedUser, nil
	}
	// otherwise, call the Google userinfo API
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v1/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+credential)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("unsuccessful response from userinfo service: %d (%s): %s",
			resp.StatusCode, resp.Status, string(b))
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
	mutex.Lock()
	users[credential] = user
	mutex.Unlock()
	log.Printf("verified user: %+v for %s", user, credential)
	return user, nil
}

// GoogleToken holds information about a Google identity token (a JWT).
type GoogleToken struct {
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
}

// in-memory cache of tokens
var tokens map[string]*GoogleToken

func getVerifiedToken(credential string) (*GoogleToken, error) {
	if tokens == nil {
		tokens = make(map[string]*GoogleToken)
	}
	// first check the cache
	cachedToken := tokens[credential]
	if cachedToken != nil {
		log.Printf("cached token: %+v for %s", cachedToken, credential)
		return cachedToken, nil
	}
	// otherwise, call the Google tokeninfo API
	req, err := http.NewRequest("GET", "https://oauth2.googleapis.com/tokeninfo", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("id_token", credential)
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("unsuccessful response from tokeninfo service: %d (%s): %s",
			resp.StatusCode, resp.Status, string(b))
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	token := &GoogleToken{}
	err = json.Unmarshal(b, token)
	if err != nil {
		return nil, err
	}
	tokens[credential] = token
	log.Printf("verified token: %+v for %s", token, credential)
	return token, nil
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

func denyMalformedCredentials() *auth.CheckResponse {
	return &auth.CheckResponse{
		Status: &rpcstatus.Status{
			Code: int32(rpc.UNAUTHENTICATED),
		},
		HttpResponse: &auth.CheckResponse_DeniedResponse{
			DeniedResponse: &auth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{
					Code: envoy_type.StatusCode_Unauthorized,
				},
				Body: "Authorization header is malformed",
			},
		},
	}
}

func main() {
	flag.Parse()

	if *configFlag != "" {
		b, err := ioutil.ReadFile(*configFlag)
		if err != nil {
			log.Fatalf("Failed to read file: %s", err)
		}
		b = []byte(os.ExpandEnv(string(b)))
		if err := yaml.Unmarshal(b, &config); err != nil {
			log.Fatalf("Failed to unmarshal yaml: %s", err)
		}
	} else {
		// if no configuration is specified, allow all authenticated users to read and write.
		config.Readers = []string{"*"}
		config.Writers = []string{"*"}
	}

	// marshal and print current configuration for logging
	configJSON, _ := json.Marshal(config)
	log.Printf("authz-server %s", configJSON)

	lis, err := net.Listen("tcp", *portFlag)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{grpc.MaxConcurrentStreams(10)}
	s := grpc.NewServer(opts...)
	auth.RegisterAuthorizationServer(s, &authorizationServer{})
	healthpb.RegisterHealthServer(s, &healthServer{})

	log.Printf("authz-server listening on %s", *portFlag)
	_ = s.Serve(lis)
}
