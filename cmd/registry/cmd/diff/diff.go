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
	"sort"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare resources in the registry",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
			}
			for i := range args {
				args[i] = c.FQName(args[i])
			}

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			var spec1, spec2 *rpc.ApiSpec
			var path1 names.Spec
			if name1, err := names.ParseSpec(args[0]); err == nil {
				err = core.GetSpec(ctx, client, name1, true, func(s *rpc.ApiSpec) error {
					spec1 = s
					return nil
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to compare resources")
				}
				path1 = name1
			} else if name1, err := names.ParseSpecRevision(args[0]); err == nil {
				err = core.GetSpecRevision(ctx, client, name1, true, func(s *rpc.ApiSpec) error {
					spec1 = s
					return nil
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to compare resources")
				}
				path1 = name1.Spec()
			}

			if name2, err := names.ParseSpec(args[1]); err == nil {
				err = core.GetSpec(ctx, client, name2, true, func(s *rpc.ApiSpec) error {
					spec2 = s
					return nil
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to compare resources")
				}
			} else if name2, err := names.ParseSpecRevision(args[1]); err == nil {
				err = core.GetSpecRevision(ctx, client, name2, true, func(s *rpc.ApiSpec) error {
					spec2 = s
					return nil
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to compare resources")
				}
			} else if name2, err := resolveSpecRevision(ctx, client, path1.String(), args[1]); err == nil {
				err = core.GetSpecRevision(ctx, client, name2, true, func(s *rpc.ApiSpec) error {
					spec2 = s
					return nil
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to compare resources")
				}
			} else {
				log.FromContext(ctx).WithError(err).Fatal("Failed to match or handle command")
			}
			if spec1 != nil && spec2 != nil {
				err = printDiff(spec1, spec2)
			}
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to match or handle command")
			}
		},
	}
	return cmd
}

func resolveSpecRevision(ctx context.Context,
	client connection.RegistryClient,
	base string,
	suffix string) (names.SpecRevision, error) {
	// First try to treat the raw suffix as revision name.
	name, err := names.ParseSpecRevision(base + suffix)
	if err == nil {
		return name, nil
	}
	// Next try to process a relative reference '@{-N}',
	// which points to the Nth revision back.
	// NOTE: This should probably be implemented in the server
	// and removed when that is available.
	var revIndex int
	if _, err := fmt.Sscanf(suffix, "@{-%d}", &revIndex); err != nil {
		return names.SpecRevision{}, fmt.Errorf("%s is not a valid revision reference", suffix)
	} else if revIndex <= 0 {
		return names.SpecRevision{}, fmt.Errorf("%d is not a valid revision index", -revIndex)
	}
	it := client.ListApiSpecRevisions(ctx,
		&rpc.ListApiSpecRevisionsRequest{
			Name: base,
		})
	for i := 0; i >= -revIndex; i -= 1 {
		spec, err := it.Next()
		if err == iterator.Done {
			return names.SpecRevision{}, fmt.Errorf("no revision exists for index %d", -revIndex)
		} else if err != nil {
			return names.SpecRevision{}, err
		}
		if i == -revIndex {
			return names.ParseSpecRevision(spec.Name)
		}
	}
	return names.SpecRevision{}, fmt.Errorf("%s is not a valid revision reference", suffix)
}

func printDiff(spec1, spec2 *rpc.ApiSpec) error {
	if spec1.MimeType != spec2.MimeType {
		return fmt.Errorf("incomparable content types (%s, %s)", spec1.MimeType, spec2.MimeType)
	}
	if strings.Contains(spec1.MimeType, "+zip") {
		// read both zip archives into a map
		map1, err := core.UnzipArchiveToMap(spec1.Contents)
		if err != nil {
			return err
		}
		map2, err := core.UnzipArchiveToMap(spec2.Contents)
		if err != nil {
			return err
		}
		keys1 := make([]string, 0, len(map1))
		for k := range map1 {
			keys1 = append(keys1, k)
		}
		sort.Strings(keys1)
		for _, k1 := range keys1 {
			if v2, ok := map2[k1]; ok {
				diff := computeDiff(
					map1[k1],
					v2,
					spec1.Name+"/"+k1,
					spec2.Name+"/"+k1,
				)
				if len(diff) > 0 {
					fmt.Println(diff)
				}
			} else {
				fmt.Printf("--- %s/%s\n", spec1.Name, k1)
			}
		}
		keys2 := make([]string, 0, len(map2))
		for k := range map2 {
			keys2 = append(keys2, k)
		}
		sort.Strings(keys2)
		for _, k2 := range keys2 {
			if _, ok := map1[k2]; !ok {
				fmt.Printf("+++ %s/%s\n", spec2.Name, k2)
			}
		}
	} else {
		diff := computeDiff(spec1.Contents, spec2.Contents, spec1.Name, spec2.Name)
		if len(diff) > 0 {
			fmt.Println(diff)
		}
	}
	return nil
}

func computeDiff(b1, b2 []byte, n1, n2 string) string {
	s1 := string(b1)
	s2 := string(b2)
	edits := myers.ComputeEdits(span.URIFromPath(n1), s1, s2)
	return fmt.Sprint(gotextdiff.ToUnified(n1, n2, s1, edits))
}
