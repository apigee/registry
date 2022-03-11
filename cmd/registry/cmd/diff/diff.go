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

package diff

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare resources in the registry",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			var spec1, spec2 *rpc.ApiSpec
			var path1 string
			if name1, err := names.ParseSpec(args[0]); err == nil {
				spec1, err = core.GetSpec(ctx, client, name1, true, nil)
				path1 = name1.String()
			} else if name1, err := names.ParseSpecRevision(args[0]); err == nil {
				spec1, err = core.GetSpecRevision(ctx, client, name1, true, nil)
				path1 = name1.Spec().String()
			}

			if name2, err := names.ParseSpec(args[1]); err == nil {
				spec2, err = core.GetSpec(ctx, client, name2, true, nil)
			} else if name2, err := names.ParseSpecRevision(args[1]); err == nil {
				spec2, err = core.GetSpecRevision(ctx, client, name2, true, nil)
			} else if name2, err := resolveSpecRevision(ctx, client, path1, args[1]); err == nil {
				spec2, err = core.GetSpecRevision(ctx, client, name2, true, nil)
			} else {
				log.FromContext(ctx).WithError(err).Fatal("Failed to match or handle command")
			}
			if spec1 != nil && spec2 != nil {
				s1 := string(spec1.Contents)
				s2 := string(spec2.Contents)
				edits := myers.ComputeEdits(span.URIFromPath(args[0]), s1, s2)
				diff := fmt.Sprint(gotextdiff.ToUnified(args[0], args[1], s1, edits))
				fmt.Println(diff)
			}
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to match or handle command")
			}
		},
	}
	return cmd
}

func resolveSpecRevision(ctx context.Context,
	client connection.Client,
	base string,
	suffix string) (names.SpecRevision, error) {
	// First try to treat the raw suffix as revision name.
	name, err := names.ParseSpecRevision(fmt.Sprintf("%s%s", base, suffix))
	if err == nil {
		return name, nil
	}
	// Next try to process a relative reference '@{-N}',
	// which points to the Nth revision back.
	// NOTE: This should probably be implemented in the server
	// and removed when that is available.
	re := regexp.MustCompile(`@\{\-(\d+)\}`)
	m := re.FindStringSubmatch(suffix)
	if m != nil {
		it := client.ListApiSpecRevisions(ctx,
			&rpc.ListApiSpecRevisionsRequest{
				Name: base,
			})
		i, err := strconv.Atoi(m[1])
		if err != nil {
			return names.SpecRevision{}, fmt.Errorf("%s is not a valid revision reference", suffix)
		}
		for ; i >= 0; i -= 1 {
			spec, err := it.Next()
			if err == iterator.Done {
				break
			} else if err != nil {
				return names.SpecRevision{}, err
			}
			if i == 0 {
				n, err := names.ParseSpecRevision(spec.Name)
				if err != nil {
					return names.SpecRevision{}, err
				}
				return n, nil
			}
		}
		return names.SpecRevision{}, fmt.Errorf("%s is not a valid revision reference", suffix)
	}
	return names.SpecRevision{}, fmt.Errorf("%s is not a valid revision reference", suffix)
}
