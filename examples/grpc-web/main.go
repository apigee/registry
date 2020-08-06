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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/apigee/registry/rpc"
	"github.com/golang/protobuf/proto"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
)

func check(err error) {
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(-1)
	}
}

// Do ...
func Do(service, method string, req, res proto.Message) error {
	server := os.Getenv("APG_REGISTRY_AUDIENCES")
	if server == "" {
		err := fmt.Errorf("APG_REGISTRY_AUDIENCES is unset")
		return err
	}
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	buf.WriteByte(0x0)
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
	request.Header.Set("content-type", "application/grpc-web+proto")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	resBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	payloadLength := resBody[4]
	payloadBody := resBody[5 : 5+payloadLength]
	if response.StatusCode != http.StatusOK {
		stpb := &statuspb.Status{}
		err := proto.Unmarshal(payloadBody, stpb)
		if err != nil {
			return err
		}
		st := status.FromProto(stpb)
		log.Printf("%+v", st.Err())
	}
	err = proto.Unmarshal(payloadBody, res)
	if err != nil {
		return err
	}
	return err
}

func main() {
	// Make a sample API call with gRPC Web.
	req := &rpc.ListProjectsRequest{
		PageSize: 50,
	}
	res := &rpc.ListProjectsResponse{}
	err := Do("google.cloud.apigee.registry.v1alpha1.Registry", "ListProjects", req, res)
	check(err)
	log.Printf("%+v", res)
}
