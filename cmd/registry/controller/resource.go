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
	ResourceName
	GetUpdateTimestamp() time.Time
}

type SpecResource struct {
	SpecName
	UpdateTimestamp time.Time
}

func (s SpecResource) GetUpdateTimestamp() time.Time {
	return s.UpdateTimestamp
}

type VersionResource struct {
	VersionName
	UpdateTimestamp time.Time
}

func (v VersionResource) GetUpdateTimestamp() time.Time {
	return v.UpdateTimestamp
}

type ApiResource struct {
	ApiName
	UpdateTimestamp time.Time
}

func (a ApiResource) GetUpdateTimestamp() time.Time {
	return a.UpdateTimestamp
}

type ArtifactResource struct {
	ArtifactName
	UpdateTimestamp time.Time
}

func (ar ArtifactResource) GetUpdateTimestamp() time.Time {
	return ar.UpdateTimestamp
}
