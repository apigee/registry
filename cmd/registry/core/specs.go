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
	"fmt"

	"github.com/apex/log"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
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

func GetBytesForSpec(ctx context.Context, client connection.Client, spec *rpc.ApiSpec) ([]byte, error) {
	request := &rpc.GetApiSpecContentsRequest{Name: fmt.Sprintf("%s/contents", spec.GetName())}
	contents, err := client.GetApiSpecContents(ctx, request)
	if err != nil {
		return nil, err
	}
	return contents.Data, nil
}

func UploadBytesForSpec(ctx context.Context, client connection.Client, parent, specID, style string, document proto.Message) error {
	// gzip the spec before uploading it
	messageData, _ := proto.Marshal(document)
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if _, err := zw.Write(messageData); err != nil {
		log.WithError(err).Fatal("Failed to gzip spec contents")
	}
	if err := zw.Close(); err != nil {
		log.WithError(err).Fatal("Failed to finish gzipping spec contents")
	}

	request := &rpc.UpdateApiSpecRequest{
		AllowMissing: true,
		ApiSpec: &rpc.ApiSpec{
			Name:     parent + "/specs/" + specID,
			MimeType: style,
			Contents: buf.Bytes(),
		},
	}

	if _, err := client.UpdateApiSpec(ctx, request); err != nil {
		return err
	}

	return nil
}
