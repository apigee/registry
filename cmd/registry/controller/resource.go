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
	"time"
)

// This interface is used to describe resource instances
// ResourceName is embeded, the only additional field is the UpdateTimestamp
type ResourceInstance interface {
	GetResourceName() ResourceName
	GetUpdateTimestamp() time.Time
}

type SpecResource struct {
	SpecName        ResourceName
	UpdateTimestamp time.Time
}

func (s SpecResource) GetUpdateTimestamp() time.Time {
	return s.UpdateTimestamp
}

func (s SpecResource) GetResourceName() ResourceName {
	return s.SpecName
}

type VersionResource struct {
	VersionName     ResourceName
	UpdateTimestamp time.Time
}

func (v VersionResource) GetUpdateTimestamp() time.Time {
	return v.UpdateTimestamp
}

func (v VersionResource) GetResourceName() ResourceName {
	return v.VersionName
}

type ApiResource struct {
	ApiName         ResourceName
	UpdateTimestamp time.Time
}

func (a ApiResource) GetUpdateTimestamp() time.Time {
	return a.UpdateTimestamp
}

func (a ApiResource) GetResourceName() ResourceName {
	return a.ApiName
}

type ArtifactResource struct {
	ArtifactName    ResourceName
	UpdateTimestamp time.Time
}

func (ar ArtifactResource) GetUpdateTimestamp() time.Time {
	return ar.UpdateTimestamp
}

func (ar ArtifactResource) GetResourceName() ResourceName {
	return ar.ArtifactName
}
