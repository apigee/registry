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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/apigee/registry/cmd/registry/cmd/check/tree"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
)

type contextKey int

const (
	ContextKeyRegistryClient contextKey = iota
)

func RegistryClient(ctx context.Context) connection.RegistryClient {
	if c := ctx.Value(ContextKeyRegistryClient); c != nil {
		return c.(connection.RegistryClient)
	}
	return nil
}

type Resource interface {
	GetName() string
}

// Checker checks API files and returns a list of detected problems.
type Checker struct {
	rules   RuleRegistry
	configs Configs
}

// New creates and returns a checker with the given rules and configs.
func New(rules RuleRegistry, configs Configs) *Checker {
	l := &Checker{
		rules:   rules,
		configs: configs,
	}
	return l
}

func (l *Checker) Check(ctx context.Context, admin connection.AdminClient, client connection.RegistryClient, root names.Name, filter string, jobs int) (response *Response, err error) {
	response = &Response{
		RunTime:  time.Now(),
		Problems: make([]Problem, 0),
	}

	// enable rules to access client
	ctx = context.WithValue(ctx, ContextKeyRegistryClient, client)
	taskQueue, wait := core.WorkerPool(ctx, jobs)
	defer func() {
		wait()
		if response.Error != nil { // from a panic
			err = response.Error
		}
	}()

	handler := &listHandler{
		taskQueue: taskQueue,
		newTask: func(r Resource) *checkTask {
			return &checkTask{l, response, r}
		},
	}
	err = tree.ListSubresources(ctx, admin, client, root, filter, true, handler)
	return response, err
}

type checkTask struct {
	checker  *Checker
	response *Response
	resource Resource
}

func (t *checkTask) String() string {
	return "check " + t.resource.GetName()
}

func (t *checkTask) Run(ctx context.Context) error {
	var problems []Problem
	var errMessages []string
	for name, rule := range t.checker.rules {
		if t.checker.configs.IsRuleEnabled(string(name), t.resource.GetName()) {
			if probs, err := t.runAndRecoverFromPanics(ctx, rule, t.resource); err == nil {
				for _, p := range probs {
					if p.RuleID == "" {
						p.RuleID = rule.GetName()
					}
					if p.Location == "" {
						p.Location = t.resource.GetName()
					}
					problems = append(problems, p)
				}
			} else {
				errMessages = append(errMessages, err.Error())
			}
		}
	}
	var err error
	if len(errMessages) != 0 {
		err = errors.New(strings.Join(errMessages, "; "))
	}

	t.response.append(t.resource, problems, err)
	return nil
}

func (c *checkTask) runAndRecoverFromPanics(ctx context.Context, rule Rule, resource Resource) (probs []Problem, err error) {
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(error); ok {
				err = rerr
			} else {
				err = fmt.Errorf("panic occurred during rule execution: %v", r)
			}
		}
	}()

	return rule.Apply(ctx, resource), nil
}

type listHandler struct {
	taskQueue chan<- core.Task
	newTask   func(r Resource) *checkTask
}

func (c *listHandler) queueTask(r Resource) error {
	c.taskQueue <- c.newTask(r)
	return nil
}

func (c *listHandler) ProjectHandler() visitor.ProjectHandler {
	return func(p *rpc.Project) error {
		return c.queueTask(p)
	}
}

func (c *listHandler) ApiHandler() visitor.ApiHandler {
	return func(a *rpc.Api) error {
		return c.queueTask(a)
	}
}
func (c *listHandler) DeploymentHandler() visitor.DeploymentHandler {
	return func(a *rpc.ApiDeployment) error {
		return c.queueTask(a)
	}
}
func (c *listHandler) VersionHandler() visitor.VersionHandler {
	return func(a *rpc.ApiVersion) error {
		return c.queueTask(a)
	}
}

func (c *listHandler) SpecHandler() visitor.SpecHandler {
	return func(a *rpc.ApiSpec) error {
		return c.queueTask(a)
	}
}

func (c *listHandler) ArtifactHandler() visitor.ArtifactHandler {
	return func(a *rpc.Artifact) error {
		return c.queueTask(a)
	}
}
