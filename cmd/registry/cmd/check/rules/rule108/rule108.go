// Copyright 2023 Google LLC.
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

package rule108

import (
	"context"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/util"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/rpc"
)

var ruleNum = 108
var ruleName = lint.NewRuleName(ruleNum, "apideployment-external-channel-uri-format")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		externalChannelUriFormat,
	)
}

var externalChannelUriFormat = &lint.ApiDeploymentRule{
	Name: ruleName,
	OnlyIf: func(a *rpc.ApiDeployment) bool {
		return a.EndpointUri != ""
	},
	ApplyToApiDeployment: func(ctx context.Context, a *rpc.ApiDeployment) []*check.Problem {
		return util.CheckURI("external_channel_uri", a.EndpointUri)
	},
}
