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

package core

import (
	"bytes"
	"compress/gzip"
	"context"
	"log"
	"strings"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	rpcpb "github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
)

func ResourceNameOfSpec(segments []string) string {
	if len(segments) == 4 {
		return "projects/" + segments[0] +
			"/apis/" + segments[1] +
			"/versions/" + segments[2] +
			"/specs/" + segments[3]
	}
	return ""
}

func GetBytesForSpec(spec *rpc.ApiSpec) ([]byte, error) {
	if strings.Contains(spec.GetMimeType(), "+gzip") {
		return GUnzippedBytes(spec.GetContents())
	} else {
		return spec.GetContents(), nil
	}
}

func UploadBytesForSpec(ctx context.Context, client connection.Client, parent string, specID string, style string, document proto.Message) error {
	// gzip the spec before uploading it
	messageData, err := proto.Marshal(document)
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	_, err = zw.Write(messageData)
	if err != nil {
		log.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}
	spec := &rpcpb.ApiSpec{}
	spec.MimeType = style
	spec.Contents = buf.Bytes()

	request := &rpc.CreateApiSpecRequest{}
	request.Parent = parent
	request.ApiSpecId = specID
	request.ApiSpec = spec
	_, err = client.CreateApiSpec(ctx, request)
	if err != nil {
		// if this fails, we should try calling UpdateSpec
		spec.Name = parent + "/specs/" + specID
		request := &rpc.UpdateApiSpecRequest{}
		request.ApiSpec = spec
		_, err = client.UpdateApiSpec(ctx, request)
		return err
	}
	return nil
}
