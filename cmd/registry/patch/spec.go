// Copyright 2022 Google LLC.
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
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/encoding"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"gopkg.in/yaml.v3"
)

// NewApiSpec allows an API spec to be individually exported as a YAML file.
func NewApiSpec(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiSpec, nested bool) (*encoding.ApiSpec, error) {
	specName, err := names.ParseSpecRevision(message.Name)
	if err != nil {
		return nil, err
	}
	var artifacts []*encoding.Artifact
	if nested {
		artifacts, err = collectChildArtifacts(ctx, client, specName.Artifact("-"))
		if err != nil {
			return nil, err
		}
	}
	return &encoding.ApiSpec{
		Header: encoding.Header{
			ApiVersion: encoding.RegistryV1,
			Kind:       "Spec",
			Metadata: encoding.Metadata{
				Name:        specName.SpecID,
				Parent:      names.ExportableName(specName.Parent(), specName.ProjectID),
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
		Data: encoding.ApiSpecData{
			FileName:    message.Filename,
			Description: message.Description,
			MimeType:    message.MimeType,
			SourceURI:   message.SourceUri,
			Artifacts:   artifacts,
		},
	}, nil
}

func applyApiSpecPatchBytes(
	ctx context.Context,
	client connection.RegistryClient,
	bytes []byte,
	project string,
	filename string) error {
	var spec encoding.ApiSpec
	err := yaml.Unmarshal(bytes, &spec)
	if err != nil {
		return err
	}
	return applyApiSpecPatch(ctx, client, &spec, project, filename)
}

func specName(parent string, metadata encoding.Metadata) (names.Spec, error) {
	if metadata.Parent != "" {
		parent = parent + "/" + metadata.Parent
	}
	version, err := names.ParseVersion(parent)
	if err != nil {
		return names.Spec{}, err
	}
	return version.Spec(metadata.Name), nil
}

func applyApiSpecPatch(
	ctx context.Context,
	client connection.RegistryClient,
	spec *encoding.ApiSpec,
	parent string,
	filename string) error {
	name, err := specName(parent, spec.Metadata)
	if err != nil {
		return err
	}
	req := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:        name.String(),
			Filename:    spec.Data.FileName,
			Description: spec.Data.Description,
			MimeType:    spec.Data.MimeType,
			SourceUri:   spec.Data.SourceURI,
			Labels:      spec.Metadata.Labels,
			Annotations: spec.Metadata.Annotations,
		},
		AllowMissing: true,
	}
	// if the spec's filename points to a local file, use that as the spec's contents
	if filename != "" {
		body, err := os.ReadFile(filepath.Join(filepath.Dir(filename), spec.Data.FileName))
		if err == nil {
			if strings.Contains(spec.Data.MimeType, "+gzip") {
				body, err = compress.GZippedBytes(body)
				if err != nil {
					return err
				}
			}
			req.ApiSpec.Contents = body
		}
	}
	// if we didn't find the spec contents in a file and it was supposed to be a zip archive,
	// create it from the contents of the directory where we found the YAML file.
	if req.ApiSpec.Contents == nil && mime.IsZipArchive(spec.Data.MimeType) {
		container := filepath.Dir(filename)
		filenames := []string{}
		err := filepath.WalkDir(container, func(p string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			} else if entry.IsDir() {
				return nil // Do nothing for the directory, but still walk its contents.
			} else if p == filename || strings.HasSuffix(p, ".zip") {
				return nil // Omit the Registry YAML file and any zip archives.
			} else {
				filenames = append(filenames, strings.TrimPrefix(p, container+"/"))
			}
			return nil
		})
		if err != nil {
			return err
		}
		buf, err := compress.ZipArchiveOfFiles(filenames, container+"/")
		if err != nil {
			return err
		}
		req.ApiSpec.Contents = buf.Bytes()
	}
	// if we didn't find the spec body in a file, try to read it from the SourceURI
	if req.ApiSpec.Contents == nil && spec.Data.SourceURI != "" {
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
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			if strings.Contains(spec.Data.MimeType, "+gzip") {
				body, err = compress.GZippedBytes(body)
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
				contents, err := compress.ZipArchiveOfPath(path, "", recursive)
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
					body, err = compress.GZippedBytes(body)
					if err != nil {
						return err
					}
				}
				req.ApiSpec.Contents = body
			}
		}
	}
	_, err = client.UpdateApiSpec(ctx, req)
	if err != nil {
		return fmt.Errorf("UpdateApiSpec: %s", err)
	}
	for _, artifactPatch := range spec.Data.Artifacts {
		err = applyArtifactPatch(ctx, client, artifactPatch, name.String(), filename)
		if err != nil {
			return err
		}
	}
	return nil
}
