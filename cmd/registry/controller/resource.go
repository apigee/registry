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
type resourceInstance interface {
	ResourceName() resourceName
	UpdateTimestamp() time.Time
}

type specResource struct {
	specName        resourceName
	updateTimestamp time.Time
}

func (s specResource) UpdateTimestamp() time.Time {
	return s.updateTimestamp
}

func (s specResource) ResourceName() resourceName {
	return s.specName
}

type versionResource struct {
	versionName     resourceName
	updateTimestamp time.Time
}

func (v versionResource) UpdateTimestamp() time.Time {
	return v.updateTimestamp
}

func (v versionResource) ResourceName() resourceName {
	return v.versionName
}

type apiResource struct {
	apiName         resourceName
	updateTimestamp time.Time
}

func (a apiResource) UpdateTimestamp() time.Time {
	return a.updateTimestamp
}

func (a apiResource) ResourceName() resourceName {
	return a.apiName
}

type artifactResource struct {
	artifactName    resourceName
	updateTimestamp time.Time
}

func (ar artifactResource) UpdateTimestamp() time.Time {
	return ar.updateTimestamp
}

func (ar artifactResource) ResourceName() resourceName {
	return ar.artifactName
}
