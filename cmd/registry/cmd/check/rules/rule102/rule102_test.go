// Copyright 2023 Google LLC. All Rights Reserved.
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

package rule102

import (
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
)

func TestAddRules(t *testing.T) {
	if err := AddRules(lint.NewRuleRegistry()); err != nil {
		t.Errorf("AddRules got an error: %v", err)
	}
}

func Test_recommendedDeploymentRef(t *testing.T) {
	for _, tt := range []struct {
		desc     string
		in       string
		expected []lint.Problem
	}{
		{"empty", "", nil},
		{"unable to parse", "bad", []lint.Problem{{
			Severity:   lint.ERROR,
			Message:    `recommended_deployment "bad" is not a valid ApiDeployment name.`,
			Suggestion: `Parse error: invalid deployment name "bad": must match "^projects/([A-Za-z0-9-.]+)/locations/global/apis/([A-Za-z0-9-.]+)/deployments/([A-Za-z0-9-.]+)$"`,
		}}},
		{"not a child", "projects/myproject/locations/global/apis/bad/deployments/bad", []lint.Problem{{
			Severity:   lint.ERROR,
			Message:    `recommended_deployment "projects/myproject/locations/global/apis/bad/deployments/bad" is not a child of this Api.`,
			Suggestion: `Correct the recommended_deployment.`,
		}}},
		{"good", "projects/myproject/locations/global/apis/myapi/deployments/good", nil},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			a := &rpc.Api{
				Name:                  "projects/myproject/locations/global/apis/myapi",
				RecommendedDeployment: tt.in,
			}

			if recommendedDeploymentRef.OnlyIf(a) {
				got := recommendedDeploymentRef.ApplyToApi(context.Background(), a)
				if diff := cmp.Diff(got, tt.expected); diff != "" {
					t.Errorf("unexpected diff: (-want +got):\n%s", diff)
				}
			}
		})
	}
}
