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

package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	rpcpb "github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
)

func init() {
	uploadCmd.AddCommand(specCmd)
	specCmd.Flags().String("version", "", "Version for uploaded spec")
	specCmd.Flags().String("style", "", "style of spec to upload (openapi, discovery, proto)")
}

var specCmd = &cobra.Command{
	Use:   "spec",
	Short: "Upload an API spec",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		flagset := cmd.LocalFlags()
		version, err := flagset.GetString("version")
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		if version == "" {
			log.Fatalf("Please specify a version")
		}
		style, err := flagset.GetString("style")
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		if style == "" {
			log.Fatal("Please specify a style")
		}
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		for _, arg := range args {
			matches, err := filepath.Glob(arg)
			if err != nil {
				log.Printf("%s\n", err.Error())
			}
			// for each match, upload the file
			for _, match := range matches {
				log.Printf("now upload %+v", match)
				fi, err := os.Stat(match)
				if err == nil {
					switch mode := fi.Mode(); {
					case mode.IsDir():
						fmt.Printf("upload directory %s\n", match)
						err = uploadSpecDirectory(match, client, version, style)
					case mode.IsRegular():
						fmt.Printf("upload file %s\n", match)
						err = uploadSpecFile(match, client, version, style)
					}
					if err != nil {
						log.Fatalf("%s", err.Error())
					}
				} else {
					log.Fatalf("%s", err.Error())
				}
			}
		}
	},
}

func uploadSpecDirectory(dirname string, client *gapic.RegistryClient, version string, style string) error {
	if style != "proto" {
		return fmt.Errorf("unsupported directory style %s", style)
	}
	prefix := dirname + "/"
	// build a zip archive with the contents of the path
	// https://golangcode.com/create-zip-files-in-go/
	buf, err := core.ZipArchiveOfPath(dirname, prefix)
	if err != nil {
		return err
	}
	ctx := context.TODO()
	request := &rpcpb.CreateApiSpecRequest{
		Parent:    version,
		ApiSpecId: "protos.zip",
		ApiSpec: &rpcpb.ApiSpec{
			MimeType: style,
			Filename: core.ProtoMimeType("+zip"),
			Contents: buf.Bytes(),
		},
	}
	response, err := client.CreateApiSpec(ctx, request)
	if err == nil {
		log.Printf("created %s", response.Name)
	} else if core.AlreadyExists(err) {
		log.Printf("found %s/specs/%s", request.Parent, request.ApiSpecId)
	} else {
		details := fmt.Sprintf("contents-length: %d", len(request.ApiSpec.Contents))
		log.Printf("error %s/specs/%s: %s [%s]",
			request.Parent, request.ApiSpecId, err.Error(), details)
	}
	return nil
}

func uploadSpecFile(filename string, client *gapic.RegistryClient, version string, style string) error {
	var mimeType string
	switch style {
	case "openapi":
		if strings.Contains(filename, "swagger") { // TODO: switch on actual spec contents
			mimeType = core.OpenAPIMimeType("+gzip", "2")
		} else {
			mimeType = core.OpenAPIMimeType("+gzip", "3")
		}
		break
	case "discovery":
		mimeType = core.DiscoveryMimeType("+gzip")
		break
	default:
		return fmt.Errorf("unsupported file style %s", style)
	}
	specID := filepath.Base(filename)
	// does the spec file exist? if not, create it
	request := &rpcpb.GetApiSpecRequest{}
	request.Name = version + "/specs/" + specID
	ctx := context.TODO()
	_, err := client.GetApiSpec(ctx, request)
	if err != nil { // TODO only do this for NotFound errors
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("err %+v", err)
		} else {
			request := &rpcpb.CreateApiSpecRequest{}
			request.Parent = version
			request.ApiSpecId = specID
			request.ApiSpec = &rpcpb.ApiSpec{}
			request.ApiSpec.Filename = specID
			request.ApiSpec.Contents, err = core.GZippedBytes(bytes)
			request.ApiSpec.MimeType = mimeType
			response, err := client.CreateApiSpec(ctx, request)
			log.Printf("response %+v\nerr %+v", response, err)
		}
	}
	return nil
}
