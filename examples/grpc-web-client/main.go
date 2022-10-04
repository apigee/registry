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
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/apigee/registry/rpc"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func check(err error) {
	if err != nil {
		log.Printf("ERROR %+v", err)
		os.Exit(-1)
	}
}

// Do performs a gRPC Web request.
func Do(server, service, method string, req, res proto.Message) error {
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	buf.WriteByte(0x0) // compressed-flag, message is uncompressed
	err = binary.Write(buf, binary.BigEndian, int32(len(reqBytes)))
	if err != nil {
		return err
	}
	buf.Write(reqBytes)
	url := server + "/" + service + "/" + method
	request, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return err
	}
	token := os.Getenv("APG_REGISTRY_TOKEN")
	request.Header.Set("authorization", "Bearer "+token)
	request.Header.Set("content-type", "application/grpc-web+proto")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("non-ok response %d", response.StatusCode)
	}
	ss := response.Header["Grpc-Status"]
	if len(ss) > 0 {
		return fmt.Errorf("call completed with gRPC status %s", ss[0])
	}
	var compression int8
	err = binary.Read(response.Body, binary.BigEndian, &compression)
	if err != nil {
		return err
	}
	if compression != 0 {
		return errors.New("unsupported compressed response")
	}
	var payloadLength int32
	err = binary.Read(response.Body, binary.BigEndian, &payloadLength)
	if err != nil {
		return err
	}
	payloadBody := make([]byte, payloadLength)
	c, err := response.Body.Read(payloadBody)
	if err != nil {
		return err
	}
	if c != int(payloadLength) {
		return fmt.Errorf("error reading payload, expected %d bytes, got %d", payloadLength, c)
	}
	if response.StatusCode != http.StatusOK {
		stpb := &statuspb.Status{}
		err := proto.Unmarshal(payloadBody, stpb)
		if err != nil {
			return err
		}
		st := status.FromProto(stpb)
		log.Printf("status %+v", st.Err())
	}
	err = proto.Unmarshal(payloadBody, res)
	if err != nil {
		return err
	}
	return err
}

func main() {
	prefix := "https://"
	if insecure, _ := strconv.ParseBool(os.Getenv("APG_REGISTRY_INSECURE")); insecure {
		prefix = "http://"
	}
	server := prefix + os.Getenv("APG_REGISTRY_ADDRESS")

	// Make a sample API call with gRPC Web.
	req := &rpc.ListProjectsRequest{
		PageSize: 4,
	}
	res := &rpc.ListProjectsResponse{}
	err := Do(server, "google.cloud.apigeeregistry.v1.Registry", "ListProjects", req, res)
	check(err)
	log.Printf("response: %+v", res)
}
