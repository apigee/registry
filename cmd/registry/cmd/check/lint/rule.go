// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lint

import (
	"context"
	"reflect"

	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/rpc"
)

// Rule defines a rule for checking a Resource.
// Anything that satisfies this interface can be used as a rule,
// but most rule authors will want to use the implementations provided.
type Rule interface {
	// GetName returns the name of the rule.
	GetName() RuleName

	// Apply accepts a resource and checks it,
	// returning a slice of Problems it finds.
	Apply(ctx context.Context, resource Resource) []*check.Problem
}

type ProjectRule struct {
	Name RuleName

	// OnlyIf determines whether this rule is applicable.
	OnlyIf func(p *rpc.Project) bool

	// ApplyToProject accepts a Project and checks it,
	// returning a slice of Problems it finds.
	ApplyToProject func(ctx context.Context, p *rpc.Project) []*check.Problem
}

// GetName returns the name of the rule.
func (r *ProjectRule) GetName() RuleName {
	return r.Name
}

// Apply calls ApplyToProject if the Resource is a Project.
func (r *ProjectRule) Apply(ctx context.Context, res Resource) (problems []*check.Problem) {
	if p, ok := res.(*rpc.Project); ok {
		if r.OnlyIf == nil || r.OnlyIf(p) {
			problems = r.ApplyToProject(ctx, p)
		}
	}

	return problems
}

type ArtifactRule struct {
	Name RuleName

	// OnlyIf determines whether this rule is applicable.
	OnlyIf func(p *rpc.Artifact) bool

	// ApplyToArtifact accepts an Artifact and checks it,
	// returning a slice of Problems it finds.
	ApplyToArtifact func(ctx context.Context, p *rpc.Artifact) []*check.Problem
}

// GetName returns the name of the rule.
func (r *ArtifactRule) GetName() RuleName {
	return r.Name
}

// Apply calls ApplyToArtifact if the Resource is an Artifact.
func (r *ArtifactRule) Apply(ctx context.Context, res Resource) (problems []*check.Problem) {
	if a, ok := res.(*rpc.Artifact); ok {
		if r.OnlyIf == nil || r.OnlyIf(a) {
			problems = r.ApplyToArtifact(ctx, a)
		}
	}
	return problems
}

type ApiRule struct {
	Name RuleName

	// OnlyIf determines whether this rule is applicable.
	OnlyIf func(p *rpc.Api) bool

	// ApplyToApi accepts an Api and checks it,
	// returning a slice of Problems it finds.
	ApplyToApi func(ctx context.Context, p *rpc.Api) []*check.Problem
}

// GetName returns the name of the rule.
func (r *ApiRule) GetName() RuleName {
	return r.Name
}

// Apply calls ApplyToApi if the Resource is an Api.
func (r *ApiRule) Apply(ctx context.Context, res Resource) (problems []*check.Problem) {
	if a, ok := res.(*rpc.Api); ok {
		if r.OnlyIf == nil || r.OnlyIf(a) {
			problems = r.ApplyToApi(ctx, a)
		}
	}
	return problems
}

type ApiDeploymentRule struct {
	Name RuleName

	// OnlyIf determines whether this rule is applicable.
	OnlyIf func(p *rpc.ApiDeployment) bool

	// ApplyToApiDeployment accepts an ApiDeployment and checks it,
	// returning a slice of Problems it finds.
	ApplyToApiDeployment func(ctx context.Context, p *rpc.ApiDeployment) []*check.Problem
}

// GetName returns the name of the rule.
func (r *ApiDeploymentRule) GetName() RuleName {
	return r.Name
}

// Apply calls ApplyToApiDeployment if the Resource is an ApiDeployment.
func (r *ApiDeploymentRule) Apply(ctx context.Context, res Resource) (problems []*check.Problem) {
	if a, ok := res.(*rpc.ApiDeployment); ok {
		if r.OnlyIf == nil || r.OnlyIf(a) {
			problems = r.ApplyToApiDeployment(ctx, a)
		}
	}
	return problems
}

type ApiVersionRule struct {
	Name RuleName

	// OnlyIf determines whether this rule is applicable.
	OnlyIf func(p *rpc.ApiVersion) bool

	// ApplyToApiVersion accepts a Version and checks it,
	// returning a slice of Problems it finds.
	ApplyToApiVersion func(ctx context.Context, p *rpc.ApiVersion) []*check.Problem
}

// GetName returns the name of the rule.
func (r *ApiVersionRule) GetName() RuleName {
	return r.Name
}

// Apply calls ApplyToVersion if the Resource is an ApiVersion.
func (r *ApiVersionRule) Apply(ctx context.Context, res Resource) (problems []*check.Problem) {
	if a, ok := res.(*rpc.ApiVersion); ok {
		if r.OnlyIf == nil || r.OnlyIf(a) {
			problems = r.ApplyToApiVersion(ctx, a)
		}
	}
	return problems
}

type ApiSpecRule struct {
	Name RuleName

	// OnlyIf determines whether this rule is applicable.
	OnlyIf func(p *rpc.ApiSpec) bool

	// ApiSpecRule accepts an ApiSpec and checks it,
	// returning a slice of Problems it finds.
	ApplyToApiSpec func(ctx context.Context, p *rpc.ApiSpec) []*check.Problem
}

// GetName returns the name of the rule.
func (r *ApiSpecRule) GetName() RuleName {
	return r.Name
}

// Apply calls ApplyToApiSpec if the Resource is an ApiSpec.
func (r *ApiSpecRule) Apply(ctx context.Context, res Resource) (problems []*check.Problem) {
	if a, ok := res.(*rpc.ApiSpec); ok {
		if r.OnlyIf == nil || r.OnlyIf(a) {
			problems = r.ApplyToApiSpec(ctx, a)
		}
	}
	return problems
}

// FieldRule defines a rule that is run on each field within a file.
type FieldRule struct {
	Name RuleName

	// OnlyIf determines whether this rule is applicable.
	OnlyIf func(resource Resource, field string) bool

	// ApplyToField accepts a Field name and value and checks it, returning a slice of
	// Problems it finds.
	ApplyToField func(ctx context.Context, resource Resource, field string, value interface{}) []*check.Problem
}

// GetName returns the name of the rule.
func (r *FieldRule) GetName() RuleName {
	return r.Name
}

// Apply visits every field in the passed resource and runs `ApplyToField`.
//
// If an `OnlyIf` function is provided on the rule, it is run against each
// field, and if it returns false, the `ApplyToField` function is not called.
func (r *FieldRule) Apply(ctx context.Context, res Resource) (problems []*check.Problem) {
	t := reflect.TypeOf(res)
	v := reflect.ValueOf(res)
	in := reflect.Indirect(v)
	for i := 0; i < in.NumField(); i++ {
		name := t.Elem().Field(i).Name
		if r.OnlyIf == nil || r.OnlyIf(res, name) {
			value := v.Elem().Field(i).Interface()
			probs := r.ApplyToField(ctx, res, name, value)
			for i := range probs {
				if probs[i].Location == "" {
					probs[i].Location = res.GetName() + "::" + name
				}
			}
			problems = append(problems, probs...)
		}
	}
	return problems
}
