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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/apex/log"
	"github.com/yoheimuta/go-protoparser/v4/parser"

	protoparser "github.com/yoheimuta/go-protoparser/v4"
	yaml "gopkg.in/yaml.v3"
)

// Details contains details about a spec.
type Details struct {
	Title    string
	Services []string
}

// NewDetailsFromZippedProtos returns a Details structure describing a spec.
func NewDetailsFromZippedProtos(b []byte) (*Details, error) {
	// create a tmp directory
	dname, err := ioutil.TempDir("", "registry-protos-")
	if err != nil {
		return nil, err
	}
	// whenever we finish, delete the tmp directory
	defer os.RemoveAll(dname)
	// unzip the protos to the temp directory
	_, err = unzipArchiveToPath(b, dname)
	if err != nil {
		return nil, err
	}
	// visit all the files in the temp directory
	details := make([]*Details, 0)
	err = filepath.Walk(dname,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".proto") {
				details = append(details, analyzeProto(path))
			}
			if strings.HasSuffix(path, ".yaml") && !strings.Contains(path, "gapic") {
				details = append(details, analyzeYaml(path))
			}
			return nil
		})
	if err != nil {
		log.WithError(err).Debug("Failed to walk directory")
	}
	if len(details) == 1 {
		return details[0], nil
	}
	if len(details) > 1 {
		summary := &Details{
			Title:    details[0].Title,
			Services: details[0].Services,
		}
		for i, d := range details {
			if i > 0 && d != nil {
				if d.Title != "" {
					summary.Title = d.Title
				}
				summary.Services = append(summary.Services, d.Services...)
			}
		}
		sort.Strings(summary.Services)
		return summary, nil
	}
	return &Details{
		Title:    "",
		Services: nil,
	}, nil
}

func analyzeYaml(filename string) *Details {
	var err error
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil
	}
	var info map[string]interface{}
	err = yaml.Unmarshal(bytes, &info)
	if err != nil {
		return nil
	}
	documentType := info["type"]
	if documentType == nil {
		return nil
	}
	if documentType.(string) != "google.api.Service" {
		return nil
	}
	title := info["title"].(string)
	return &Details{Title: title}
}

func analyzeProto(filename string) *Details {
	reader, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer reader.Close()
	p, err := protoparser.Parse(
		reader,
		protoparser.WithDebug(false),
		protoparser.WithPermissive(true),
		protoparser.WithFilename(filepath.Base(filename)),
	)
	if err != nil {
		return nil
	}
	details := &Details{}
	for _, x := range p.ProtoBody {
		switch m := x.(type) {
		case *parser.Service:
			details.Services = append(details.Services, m.ServiceName)
		default:
			// fmt.Printf("IGNORED %+v\n", v)
		}
	}
	return details
}
