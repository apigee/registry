// Copyright 2022 Google LLC. All Rights Reserved.
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

package patch

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

type ApiSpecData struct {
	FileName    string `yaml:"filename,omitempty"`
	Description string `yaml:"description,omitempty"`
	MimeType    string `yaml:"mimeType,omitempty"`
	SourceURI   string `yaml:"sourceURI,omitempty"`
}

type ApiSpec struct {
	Header `yaml:",inline"`
	Data   ApiSpecData `yaml:"data"`
}

func newApiSpec(message *rpc.ApiSpec) (*ApiSpec, error) {
	specName, err := names.ParseSpec(message.Name)
	if err != nil {
		return nil, err
	}
	return &ApiSpec{
		Header: Header{
			ApiVersion: RegistryV1,
			Kind:       "ApiSpec",
			Metadata: Metadata{
				Name:        specName.SpecID,
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
		Data: ApiSpecData{
			FileName:    message.Filename,
			Description: message.Description,
			MimeType:    message.MimeType,
			SourceURI:   message.SourceUri,
		},
	}, nil
}

func applyApiSpecPatch(
	ctx context.Context,
	client connection.Client,
	spec *ApiSpec,
	parent string) error {
	name := fmt.Sprintf("%s/specs/%s", parent, spec.Metadata.Name)
	req := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:        name,
			Filename:    spec.Data.FileName,
			Description: spec.Data.Description,
			MimeType:    spec.Data.MimeType,
			SourceUri:   spec.Data.SourceURI,
			Labels:      spec.Metadata.Labels,
			Annotations: spec.Metadata.Annotations,
		},
		AllowMissing: true,
	}
	// TODO: verify mime type
	if spec.Data.SourceURI != "" {
		u, err := url.ParseRequestURI(spec.Data.SourceURI)
		if err != nil {
			return err
		}
		switch u.Scheme {
		case "http", "https":
			resp, err := http.Get(spec.Data.SourceURI)
			if err != nil {
				return err
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			if strings.Contains(spec.Data.MimeType, "+gzip") {
				body, err = core.GZippedBytes(body)
				if err != nil {
					return err
				}
			}
			req.ApiSpec.Contents = body
		case "file":
			// Remove leading slash from path.
			// We expect to load from paths relative to the working directory,
			// but users can add an additional slash to specify a global path.
			path := strings.TrimPrefix(u.Path, "/")
			info, err := os.Stat(path)
			if err != nil {
				return err
			}
			if info.IsDir() {
				recursive := true
				d := u.Query()["recursive"]
				if len(d) > 0 {
					recursive, err = strconv.ParseBool(d[0])
					if err != nil {
						return err
					}
				}
				contents, err := core.ZipArchiveOfPath(path, "", recursive)
				if err != nil {
					return err
				}
				req.ApiSpec.Contents = contents.Bytes()
			} else {
				body, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if strings.Contains(spec.Data.MimeType, "+gzip") {
					body, err = core.GZippedBytes(body)
					if err != nil {
						return err
					}
				}
				req.ApiSpec.Contents = body
			}
		}
	}
	_, err := client.UpdateApiSpec(ctx, req)
	return err
}
