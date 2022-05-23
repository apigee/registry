// Copyright 2022 Google LLC
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

package scoring

import (
	"context"
	"fmt"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func scoreCardID(definitionID string) string {
	return fmt.Sprintf("scorecard-%s", definitionID)
}

func FetchScoreCardDefinitions(
	ctx context.Context,
	client connection.Client,
	resource patterns.ResourceName) ([]*rpc.Artifact, error) {
	defArtifacts := make([]*rpc.Artifact, 0)

	project := fmt.Sprintf("%s/locations/global", resource.Project())
	artifact, err := names.ParseArtifact(fmt.Sprintf("%s/artifacts/-", project))
	if err != nil {
		return nil, err
	}
	listFilter := fmt.Sprintf("mime_type == %q", patch.MimeTypeForKind("ScoreCardDefinition"))
	err = core.ListArtifacts(ctx, client, artifact, listFilter, true,
		func(artifact *rpc.Artifact) error {
			definition := &rpc.ScoreCardDefinition{}
			if err := proto.Unmarshal(artifact.GetContents(), definition); err != nil {
				return err
			}

			// Check if ScoreCardDefinition.TargetResource matches with the supplied resource
			err := matchResourceWithTarget(definition.GetTargetResource(), resource, project)
			if err != nil {
				return err
			}

			defArtifacts = append(defArtifacts, artifact)
			return nil
		})

	if err != nil {
		return nil, err
	}

	return defArtifacts, nil
}

func CalculateScoreCard(
	ctx context.Context,
	client connection.Client,
	defArtifact *rpc.Artifact,
	resource patterns.ResourceInstance,
	dryRun bool) error {
	project := fmt.Sprintf("%s/locations/global", resource.ResourceName().Project())

	// Extract definition
	definition := &rpc.ScoreCardDefinition{}
	if err := proto.Unmarshal(defArtifact.GetContents(), definition); err != nil {
		return err
	}

	var takeAction bool

	// Fetch the to be generated ScoreCard artifact (if present)
	artifactName := fmt.Sprintf("%s/artifacts/%s", resource.ResourceName().String(), scoreCardID(definition.GetId()))
	scoreCardArtifact, err := getArtifact(ctx, client, artifactName, false)
	if err != nil {
		// Generate ScoreCard if the ScoreCard artifact doesn't exist
		if status.Code(err) == codes.NotFound {
			takeAction = true
		} else {
			return fmt.Errorf("failed to fetch artifact %q: %s", artifactName, err)
		}
	}

	// Generate ScoreCard if the definition has been updated
	if scoreCardArtifact != nil && scoreCardArtifact.GetUpdateTime().AsTime().Before(defArtifact.GetUpdateTime().AsTime()) {
		takeAction = true
	}

	result := processScorePatterns(ctx, client, definition, resource, scoreCardArtifact, takeAction, project)
	if result.err != nil {
		return err
	}

	if result.needsUpdate {
		if dryRun {
			core.PrintMessage(result.scoreCard)
		} else {
			// upload the scoreCard to registry
			err = uploadScoreCard(ctx, client, resource, result.scoreCard)
			if err != nil {
				return err
			}
		}
	} else {
		log.Debugf(ctx, "ScoreCard %s is already up-to-date.", artifactName)
	}

	return nil
}

// Response returned after fetching all the scoreArtifacts to form a ScoreCard.
type scoreCardResult struct {
	// Represents the ScoreCard generated after fetching all the score artifacts from the score_patterns.
	scoreCard *rpc.ScoreCard
	// Represents if the final scoreCardArtifact needs an update
	// This is determined based on the timestamps of the existing scoreCardArtifact and the dependent scoreArtifacts fetched from score_patterns.
	needsUpdate bool
	// Represents the error encountered while generating the ScoreCard.
	err error
}

func processScorePatterns(
	ctx context.Context,
	client connection.Client,
	definition *rpc.ScoreCardDefinition,
	resource patterns.ResourceInstance,
	scoreCardArtifact *rpc.Artifact,
	takeAction bool,
	project string) scoreCardResult {
	var needsUpdate bool
	scoreArtifacts := make([]*rpc.Score, 0)

	for _, scorePattern := range definition.GetScorePatterns() {
		extendedPattern, err := patterns.SubstituteReferenceEntity(scorePattern, resource.ResourceName())
		if err != nil {
			return scoreCardResult{
				scoreCard:   nil,
				needsUpdate: false,
				err:         fmt.Errorf("invalid pattern %q in score_patterns: %s", scorePattern, err),
			}
		}

		// Fetch scoreArtifact
		artifact, err := getArtifact(ctx, client, extendedPattern.String(), true)
		if err != nil {
			return scoreCardResult{
				scoreCard:   nil,
				needsUpdate: false,
				err:         fmt.Errorf("failed to fetch artifact %s: %s", extendedPattern.String(), err),
			}
		}

		// needsUpdate tells the calling function if the ScoreCard artifact needs to be updated
		needsUpdate = needsUpdate || takeAction || scoreCardArtifact.GetUpdateTime().AsTime().Before(artifact.GetUpdateTime().AsTime())
		// Extract Score from the fetched artifact
		score := &rpc.Score{}
		if err := proto.Unmarshal(artifact.GetContents(), score); err != nil {
			return scoreCardResult{
				scoreCard:   nil,
				needsUpdate: false,
				err:         fmt.Errorf("failed unmarshaling artifact %q into Score proto: %s", artifact.GetName(), err),
			}
		}

		scoreArtifacts = append(scoreArtifacts, score)
	}

	if needsUpdate {
		// Build the final ScoreCard proto
		scoreCard := &rpc.ScoreCard{
			Id:             scoreCardID(definition.GetId()),
			Kind:           "ScoreCard",
			DisplayName:    definition.GetDisplayName(),
			Description:    definition.GetDescription(),
			DefinitionName: fmt.Sprintf("%s/artifacts/%s", project, definition.GetId()),
			Scores:         scoreArtifacts,
		}

		return scoreCardResult{
			scoreCard:   scoreCard,
			needsUpdate: true,
			err:         nil,
		}
	}

	return scoreCardResult{
		scoreCard:   nil,
		needsUpdate: false,
		err:         nil,
	}
}

func uploadScoreCard(ctx context.Context, client connection.Client, resource patterns.ResourceInstance, scoreCard *rpc.ScoreCard) error {
	artifactBytes, err := proto.Marshal(scoreCard)
	if err != nil {
		return err
	}
	artifact := &rpc.Artifact{
		Name:     fmt.Sprintf("%s/artifacts/%s", resource.ResourceName().String(), scoreCard.GetId()),
		Contents: artifactBytes,
		MimeType: patch.MimeTypeForKind("Score"),
	}
	log.Debugf(ctx, "Uploading %s", artifact.GetName())
	if err = core.SetArtifact(ctx, client, artifact); err != nil {
		return fmt.Errorf("failed to save artifact %s: %s", artifact.GetName(), err)
	}

	return nil
}
