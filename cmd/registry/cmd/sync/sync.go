// Copyright 2022 Google LLC. All Rights Reserved.
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

package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/api/apigee/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "sync ACCESS_TOKEN",
		Short: "",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var (
				ctx   = cmd.Context()
				orgs  = []string{args[0]}
				token = args[1]
			)

			for _, org := range orgs {
				env, err := newEnvMap(ctx, org)
				if err != nil {
					log.Warnf(ctx, "Failed to get hostnames for environments in %s: %s", org, err)
					continue
				}

				deps, err := deployments(ctx, org)
				if err != nil {
					log.Warnf(ctx, "Failed to list deployments for %s: %s", org, err)
					continue
				}

				for _, dep := range deps {
					name := fmt.Sprintf("%s/apis/%s/revisions/%s", org, dep.ApiProxy, dep.Revision)
					rev, err := revision(ctx, name, token)
					if err != nil {
						log.Warnf(ctx, "Failed to get revision for %s: %s", name, err)
						continue
					}

					hostnames, ok := env.Hostnames(dep.Environment)
					if !ok {
						log.Warnf(ctx, "Failed to find hostnames for environment %s", dep.Environment)
						continue
					}

					for _, hostname := range hostnames {
						envgroup, ok := env.Envgroup(hostname)
						if !ok {
							log.Warnf(ctx, "Failed to determine envgroup for hostname %q", hostname)
						}

						bytes, err := protojson.MarshalOptions{Indent: "\t"}.Marshal(&rpc.ApiDeployment{
							Name:        fmt.Sprintf("projects/my-project/locations/global/apis/%s/deployments/%s", clean(rev.ProxyName), clean(hostname)),
							DisplayName: rev.ProxyName,
							Description: rev.Description,
							EndpointUri: hostname,
							Labels: map[string]string{
								"apigeex-proxy":       name,
								"apigeex-environment": fmt.Sprintf("%s/environments/%s", org, dep.Environment),
								"apigeex-envgroup":    envgroup,
							},
						})

						if err != nil {
							log.Errorf(ctx, "Failed to marshal JSON for deployment: %s", err)
							continue
						}

						fmt.Println(string(bytes))
					}
				}
			}
		},
	}
}

func clean(s string) string {
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, ".", "-")
	return strings.ToLower(s)
}

func deployments(ctx context.Context, org string) ([]*apigee.GoogleCloudApigeeV1Deployment, error) {
	apg, err := apigee.NewService(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := apg.Organizations.Deployments.List(org).Context(ctx).Do()
	return resp.Deployments, err
}

type apiProxyRevision struct {
	Basepaths   []string `json:"basepaths,omitempty"`
	ProxyName   string   `json:"name,omitempty"`
	RevisionID  string   `json:"revision,omitempty"`
	Description string   `json:"description,omitempty"`
}

func revision(ctx context.Context, revision, token string) (*apiProxyRevision, error) {
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "https",
			Host:   "apigee.googleapis.com",
			Path:   fmt.Sprintf("/v1/%s", revision),
		},
		Header: http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", token)},
		},
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rev := &apiProxyRevision{}
	return rev, json.NewDecoder(resp.Body).Decode(rev)
}

func envgroups(ctx context.Context, org string) ([]*apigee.GoogleCloudApigeeV1EnvironmentGroup, error) {
	apg, err := apigee.NewService(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := apg.Organizations.Envgroups.List(org).Context(ctx).Do()
	return resp.EnvironmentGroups, err
}

func attachments(ctx context.Context, group string) ([]*apigee.GoogleCloudApigeeV1EnvironmentGroupAttachment, error) {
	apg, err := apigee.NewService(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := apg.Organizations.Envgroups.Attachments.List(group).Context(ctx).Do()
	return resp.EnvironmentGroupAttachments, err
}

type envMap struct {
	hostnames map[string][]string
	envgroup  map[string]string
}

func (m *envMap) Hostnames(env string) ([]string, bool) {
	if m.hostnames == nil {
		return nil, false
	}

	v, ok := m.hostnames[env]
	return v, ok
}

func (m *envMap) Envgroup(hostname string) (string, bool) {
	if m.envgroup == nil {
		return "", false
	}

	v, ok := m.envgroup[hostname]
	return v, ok
}

func newEnvMap(ctx context.Context, org string) (*envMap, error) {
	groups, err := envgroups(ctx, org)
	if err != nil {
		return nil, err
	}

	m := &envMap{
		hostnames: make(map[string][]string),
		envgroup:  make(map[string]string),
	}

	for _, group := range groups {
		envgroup := fmt.Sprintf("%s/envgroups/%s", org, group.Name)
		attachments, err := attachments(ctx, envgroup)
		if err != nil {
			return nil, err
		}

		for _, attachment := range attachments {
			for _, hostname := range group.Hostnames {
				m.hostnames[attachment.Environment] = append(m.hostnames[attachment.Environment], hostname)
				m.envgroup[hostname] = envgroup
			}
		}
	}

	return m, nil
}

func bundle(ctx context.Context, revision, token string) ([]byte, error) {
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "https",
			Host:   "apigee.googleapis.com",
			Path:   fmt.Sprintf("/v1/%s", revision),
		},
		Header: http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", token)},
		},
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
