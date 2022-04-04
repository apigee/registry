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

package upload

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func specCommand(ctx context.Context) *cobra.Command {
	var (
		version string
		style   string
	)

	cmd := &cobra.Command{
		Use:   "spec",
		Short: "Upload an API spec",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			for _, arg := range args {
				matches, err := filepath.Glob(arg)
				if err != nil {
					log.FromContext(ctx).WithError(err).Debug("Failed to glob argument")
				}
				// for each match, upload the file
				for _, match := range matches {
					log.Debugf(ctx, "Now upload %+v", match)
					fi, err := os.Stat(match)
					if err != nil {
						log.FromContext(ctx).WithError(err).Fatal("Failed to get file info")
					}

					switch mode := fi.Mode(); {
					case mode.IsDir():
						log.Debugf(ctx, "Upload directory %s", match)
						err = uploadSpecDirectory(ctx, match, client, version, style)
					case mode.IsRegular():
						log.Debugf(ctx, "Upload file %s", match)
						err = uploadSpecFile(ctx, match, client, version, style)
					}
					if err != nil {
						log.FromContext(ctx).WithError(err).Fatal("Failed to upload")
					}
				}
			}
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Version to use as parent for the spec upload")
	_ = cmd.MarkFlagRequired("version")
	cmd.Flags().StringVar(&style, "style", "", "Style of spec to upload (openapi|discovery|proto)")
	_ = cmd.MarkFlagRequired("style")
	return cmd
}

func uploadSpecDirectory(ctx context.Context, dirname string, client *gapic.RegistryClient, version string, style string) error {
	if style != "proto" {
		return fmt.Errorf("unsupported directory style %s", style)
	}
	prefix := dirname + "/"
	// build a zip archive with the contents of the path
	// https://golangcode.com/create-zip-files-in-go/
	buf, err := core.ZipArchiveOfPath(dirname, prefix, true)
	if err != nil {
		return err
	}
	request := &rpc.CreateApiSpecRequest{
		Parent:    version,
		ApiSpecId: "protos.zip",
		ApiSpec: &rpc.ApiSpec{
			MimeType: core.ProtobufMimeType("+zip"),
			Filename: core.ProtobufMimeType("+zip"),
			Contents: buf.Bytes(),
		},
	}
	response, err := client.CreateApiSpec(ctx, request)
	if err == nil {
		log.Debugf(ctx, "Created %s", response.Name)
	} else if status.Code(err) == codes.AlreadyExists {
		log.Debugf(ctx, "Found %s/specs/%s", request.Parent, request.ApiSpecId)
	} else {
		log.FromContext(ctx).WithError(err).Debugf("Error %s/specs/%s [contents-length %d]", request.Parent, request.ApiSpecId, len(request.ApiSpec.Contents))
	}
	return nil
}

func uploadSpecFile(ctx context.Context, filename string, client *gapic.RegistryClient, version string, style string) error {
	var mimeType string
	switch style {
	case "openapi":
		if strings.Contains(filename, "swagger") { // TODO: switch on actual spec contents
			mimeType = core.OpenAPIMimeType("+gzip", "2")
		} else {
			mimeType = core.OpenAPIMimeType("+gzip", "3")
		}
	case "discovery":
		mimeType = core.DiscoveryMimeType("+gzip")
	default:
		return fmt.Errorf("unsupported file style %s", style)
	}
	specID := filepath.Base(filename)
	// does the spec file exist? if not, create it
	request := &rpc.GetApiSpecRequest{}
	request.Name = version + "/specs/" + specID
	_, err := client.GetApiSpec(ctx, request)
	if err != nil { // TODO only do this for NotFound errors
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			log.FromContext(ctx).WithError(err).Debug("Failed to read file")
		} else {
			request := &rpc.CreateApiSpecRequest{
				Parent:    version,
				ApiSpecId: specID,
				ApiSpec: &rpc.ApiSpec{
					Filename: specID,
					MimeType: mimeType,
				},
			}
			request.ApiSpec.Contents, err = core.GZippedBytes(bytes)
			if err != nil {
				log.FromContext(ctx).WithError(err).Debug("Failed to compress spec contents")
			}

			response, err := client.CreateApiSpec(ctx, request)
			if err != nil {
				log.FromContext(ctx).WithError(err).Debug("Failed to create spec")
			}

			log.Debugf(ctx, "Response %+v", response)
		}
	}
	return nil
}
