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
	"archive/zip"
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/parser"

	protoparser "github.com/yoheimuta/go-protoparser/v4"
	yaml "gopkg.in/yaml.v3"
)

// The API Service Configuration contains important API properties.
type ServiceConfig struct {
	Type  string `yaml:"type"`
	Title string `yaml:"title"`
}

// Details contains details about a spec.
type Details struct {
	Title    string
	Services []string
}

// NewDetailsFromZippedProtos returns a Details structure describing a spec.
func NewDetailsFromZippedProtos(ctx context.Context, b []byte) (*Details, error) {
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}

	return &Details{
		Title:    protoTitle(r.File),
		Services: protoServices(r.File),
	}, nil
}

func protoTitle(files []*zip.File) string {
	for _, f := range files {
		if !strings.HasSuffix(f.Name, ".yaml") || strings.Contains(f.Name, "gapic") {
			continue
		}

		r, err := f.Open()
		if err != nil {
			continue
		}
		defer r.Close()

		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			continue
		}

		info := new(ServiceConfig)
		if err := yaml.Unmarshal(bytes, &info); err != nil {
			continue
		}

		if info.Type == "google.api.Service" {
			return info.Title
		}
	}

	return ""
}

func protoServices(files []*zip.File) []string {
	services := make([]string, 0)
	for _, f := range files {
		if !strings.HasSuffix(f.Name, ".proto") {
			continue
		}

		r, err := f.Open()
		if err != nil {
			continue
		}
		defer r.Close()

		opts := []protoparser.Option{
			protoparser.WithDebug(false),
			protoparser.WithPermissive(true),
			protoparser.WithFilename(filepath.Base(f.Name)),
		}

		p, err := protoparser.Parse(r, opts...)
		if err != nil {
			continue
		}

		for _, x := range p.ProtoBody {
			if m, ok := x.(*parser.Service); ok {
				services = append(services, m.ServiceName)
			}
		}
	}

	sort.Strings(services)
	return services
}
