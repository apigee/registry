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

package apply

import (
	"context"
	"errors"
	"io/fs"

	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	var fileName string
	var parent string
	var recursive bool
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply patches that add content to the API Registry",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			err = patch.Apply(ctx, client, fileName, parent, recursive)
			if errors.Is(err, fs.ErrNotExist) {
				log.FromContext(ctx).WithError(err).Fatalf("File %q doesn't exist", fileName)
			} else if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Unknown error")
			}
		},
	}
	cmd.Flags().StringVarP(&fileName, "file", "f", "", "File or directory containing the patch(es) to apply")
	cmd.Flags().StringVar(&parent, "parent", "", "Parent resource for the patch")
	cmd.Flags().BoolVarP(&recursive, "recursive", "R", false,
		"Process the directory used in -f, --file recursively. Useful when you want to manage related manifests organized within the same directory")
	return cmd
}
