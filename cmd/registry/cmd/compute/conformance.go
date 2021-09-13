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

package compute

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/apigee/registry/cmd/registry/cmd/compute/conformance"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func conformantsCommand(ctx context.Context) *cobra.Command {
	var linter string
	cmd := &cobra.Command{
		Use:   "conformance",
		Short: "Compute lint results for API specs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.Fatalf("Failed to get filter from flags: %s", err)
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}

			name := args[0]

			// Ensure that the provided argument is a spec
			var specSegments []string
			if specSegments = names.SpecRegexp().FindStringSubmatch(name); specSegments == nil {
				log.Fatalf(
					fmt.Sprintf(
						"The provided argument %s does not match the regex of a spec",
						name,
					),
				)
			}
			spec, err := names.ParseSpec(name)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}

			projectSegments := []string{"projects", spec.ProjectID}

			linterNameToLinter := make(map[string]conformance.Linter)
			err = core.ListArtifacts(ctx, client, projectSegments, filter, true, func(artifact *rpc.Artifact) {
				// Only consider artifacts which have the styleguide mimetype
				messageType, err := core.MessageTypeForMimeType(artifact.GetMimeType())
				if err != nil {
					return
				}
				if messageType != "google.cloud.apigee.registry.applications.v1alpha1.styleguide" {
					return
				}

				// Unmarshal the contents of the artifact into a style guide
				styleGuide := &rpc.StyleGuide{}
				err = proto.Unmarshal(artifact.GetContents(), styleGuide)
				if err != nil {
					return
				}

				// Construct a mapping between the linter name and the linter, and populate
				// all the rules that the linter should support
				for _, guideline := range styleGuide.GetGuidelines() {
					for _, rule := range guideline.GetRules() {
						linter_name := rule.GetLinter()
						if _, ok := linterNameToLinter[linter_name]; !ok {
							linter, err := conformance.CreateLinter(rule.GetLinter())
							if err != nil {
								// If the linter is unsupported, there is no reason
								// to prematurely exit. We can just ignore this specific
								// linter and log the message to the user.
								log.Printf("%s", err.Error())
							}
							linterNameToLinter[linter_name] = linter
						}

						linter := linterNameToLinter[linter_name]
						// Create a set of mime type
						for _, allowedMimeType := range styleGuide.MimeTypes {
							linter.AddRule(allowedMimeType, rule.GetLinterRulename())
						}
					}

				}
			})

			if err != nil {
				log.Fatalf("%s", err.Error())
			}

			// Initialize task queue.
			taskQueue, wait := core.WorkerPool(ctx, 16)
			defer wait()

			// Generate tasks.
			err = core.ListSpecs(ctx, client, specSegments, filter, func(spec *rpc.ApiSpec) {
				// Lint with every linter that supports the spec's mime type
				for _, linter := range linterNameToLinter {
					if linter.SupportsMimeType(spec.GetMimeType()) {
						taskQueue <- &computeConformantTask{
							client: client,
							spec:   spec,
							linter: linter,
						}
					}
				}
			})
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
		},
	}

	cmd.Flags().StringVar(&linter, "linter", "", "The linter to use (aip|spectral|gnostic)")
	return cmd
}

type computeConformantTask struct {
	client connection.Client
	spec   *rpc.ApiSpec
	linter conformance.Linter
}

func (task *computeConformantTask) String() string {
	return fmt.Sprintf("compute %s/conformance-%s", task.spec.GetName(), task.linter.GetName())
}

func conformanceRelation(linter string) string {
	return "conformance-" + linter
}

func (task *computeConformantTask) Run(ctx context.Context) error {
	// Get the linter
	linter := task.linter
	if linter == nil {
		return errors.New("Linter is nil")
	}

	// Get the spec's bytes
	data, err := core.GetBytesForSpec(ctx, task.client, task.spec)
	if err != nil {
		return err
	}

	// Put the spec in a temporary directory
	root, err := ioutil.TempDir("", "registry-openapi-")
	if err != nil {
		return err
	}
	name := filepath.Base(task.spec.GetName())

	// Defer the deletion of the the temporary directory
	defer os.RemoveAll(root)

	// Write the file to the temporary directory
	err = ioutil.WriteFile(filepath.Join(root, name), data, 0644)
	if err != nil {
		return err
	}

	// Lint the directory containing the spec
	lint, err := linter.Lint(task.spec.GetMimeType(), root)
	if err != nil {
		return err
	}

	// Store the Lint results as an artifact on the spec
	subject := task.spec.GetName()
	messageData, err := proto.Marshal(lint)
	if err != nil {
		return err
	}
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + conformanceRelation(task.linter.GetName()),
		MimeType: core.MimeTypeForMessageType("google.cloud.apigee.registry.applications.v1alpha1.Lint"),
		Contents: messageData,
	}
	err = core.SetArtifact(ctx, task.client, artifact)
	if err != nil {
		return err
	}

	return nil
}
