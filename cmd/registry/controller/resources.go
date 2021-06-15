// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"github.com/apigee/registry/rpc"
	"github.com/golang/protobuf/ptypes"
	"time"
	"regexp"
	"fmt"
)

type Resource interface {
	GetArtifact() string
	GetSpec() string
	GetVersion() string
	GetApi() string
	GetName() string
	GetUpdateTimestamp() time.Time
	ExtractResourceGroup (string) string
}

type SpecResource struct {
	Spec *rpc.ApiSpec

}

func (s SpecResource) GetArtifact() string {
	return ""
}

func (s SpecResource) GetSpec() string {
	return s.Spec.Name
}

func (s SpecResource) GetVersion() string {
	version := extractGroup(s.Spec.Name, "versions")
	return version
}

func (s SpecResource) GetApi() string {
	api := extractGroup(s.Spec.Name, "apis")
	return api
}

func (s SpecResource) GetName() string {
	return s.Spec.Name
}

func (s SpecResource) GetUpdateTimestamp() time.Time {
	ts, _ := ptypes.Timestamp(s.Spec.RevisionUpdateTime)
	return ts
}

func (s SpecResource) ExtractResourceGroup (group_id string) string {
	group_v := extractGroup(s.Spec.Name, group_id)
	return group_v
}

type ApiResource struct {
	Api *rpc.Api
}

func (a ApiResource) GetArtifact() string {
	return ""
}

func (a ApiResource) GetSpec() string {
	return ""
}

func (a ApiResource) GetVersion() string {
	return ""
}

func (a ApiResource) GetApi() string {
	return a.Api.Name
}

func (a ApiResource) GetName() string {
	return a.Api.Name
}

func (a ApiResource) GetUpdateTimestamp() time.Time {
	ts, _ := ptypes.Timestamp(a.Api.UpdateTime)
	return ts
}

func (a ApiResource) ExtractResourceGroup (group_id string) string {
	group_v := extractGroup(a.Api.Name, group_id)
	return group_v
}


type ArtifactResource struct {
	Artifact *rpc.Artifact
}

func (ar ArtifactResource) GetArtifact() string {
	return ar.Artifact.Name
}

func (ar ArtifactResource) GetSpec() string {
	spec := extractGroup(ar.Artifact.Name, "specs")
	return spec
}

func (ar ArtifactResource) GetVersion() string {
	version := extractGroup(ar.Artifact.Name, "versions")
	return version
}

func (ar ArtifactResource) GetApi() string {
	api := extractGroup(ar.Artifact.Name, "apis")
	return api
}

func (ar ArtifactResource) GetName() string {
	return ar.Artifact.Name
}

func (ar ArtifactResource) GetUpdateTimestamp() time.Time {
	ts, _ := ptypes.Timestamp(ar.Artifact.UpdateTime)
	return ts
}

func (ar ArtifactResource) ExtractResourceGroup (group_id string) string {
	group_v := extractGroup(ar.Artifact.Name, group_id)
	return group_v
}

func extractGroup(name string, group_name string) string {
	re := regexp.MustCompile(fmt.Sprintf(".*\\/%s\\/[^\\/]*", group_name))
	group := re.FindString(name)
	return group
}


