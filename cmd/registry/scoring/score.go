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
	"google.golang.org/protobuf/proto"
)

func FetchScoreDefinitions(
	ctx context.Context,
	client connection.Client,
	resource patterns.ResourceName) ([]*rpc.ScoreDefinition, error) {
	definitions := make([]*rpc.ScoreDefinition, 0)

	project := fmt.Sprintf("%s/locations/global", resource.Project())
	artifact, err := names.ParseArtifact(fmt.Sprintf("%s/artifacts/-", project))
	if err != nil {
		return nil, err
	}
	listFilter := fmt.Sprintf("mime_type == %q", patch.ScoreDefinitionMimeType)
	err = core.ListArtifacts(ctx, client, artifact, listFilter, true,
		func(artifact *rpc.Artifact) error {

			definition := &rpc.ScoreDefinition{}
			if err := proto.Unmarshal(artifact.GetContents(), definition); err != nil {
				return err
			}

			// Check if ScoreDefinition.TargetResource matches with the supplied resource
			err := matchResourceWithTarget(definition.GetTargetResource(), resource, project)
			if err != nil {
				return err
			}

			definitions = append(definitions, definition)
			return nil
		})

	if err != nil {
		return nil, err
	}

	return definitions, nil
}

func matchResourceWithTarget(targetPattern *rpc.ResourcePattern, resourceName patterns.ResourceName, project string) error {
	targetPatternName, err := patterns.ParseResourcePattern(fmt.Sprintf("%s/%s", project, targetPattern.GetPattern()))
	if err != nil {
		return err
	}

	// Check if targetPattern and resourceName match in type (api/version/spec)
	if fmt.Sprintf("%T", targetPatternName) != fmt.Sprintf("%T", resourceName) {
		return fmt.Errorf("resource %s doesn't match target pattern %v", resourceName.String(), targetPattern)
	}

	// Check if the individual entities match
	switch targetPatternName.(type) {
	case patterns.SpecName:
		// type casting because we need access to "Name" field of patterns.SpecName to compare specific IDs
		specPattern := targetPatternName.(patterns.SpecName)
		specResource := resourceName.(patterns.SpecName)

		if specPattern.Name.ApiID != "-" && specPattern.Name.ApiID != specResource.Name.ApiID {
			return fmt.Errorf("api mismatch in resource %s and target pattern %v", resourceName.String(), targetPattern)
		}
		if specPattern.Name.VersionID != "-" && specPattern.Name.VersionID != specResource.Name.VersionID {
			return fmt.Errorf("version mismatch in resource %s and target pattern %v", resourceName.String(), targetPattern)
		}
		if specPattern.Name.SpecID != "-" && specPattern.Name.SpecID != specResource.Name.SpecID {
			return fmt.Errorf("spec mismatch in resource %s and target pattern %v", resourceName.String(), targetPattern)
		}
	case patterns.VersionName:
		// type casting because we need access to "Name" field of patterns.VersionName to compare specific IDs
		versionPattern := targetPatternName.(patterns.VersionName)
		versionResource := resourceName.(patterns.VersionName)

		if versionPattern.Name.ApiID != "-" && versionPattern.Name.ApiID != versionResource.Name.ApiID {
			return fmt.Errorf("api mismatch in resource %s and target pattern %v", resourceName.String(), targetPattern)
		}
		if versionPattern.Name.VersionID != "-" && versionPattern.Name.VersionID != versionResource.Name.VersionID {
			return fmt.Errorf("version mismatch in resource %s and target pattern %v", resourceName.String(), targetPattern)
		}
	case patterns.ApiName:
		// type casting because we need access to "Name" field of patterns.ApiName to compare specific IDs
		apiPattern := targetPatternName.(patterns.ApiName)
		apiResource := resourceName.(patterns.ApiName)

		if apiPattern.Name.ApiID != "-" && apiPattern.Name.ApiID != apiResource.Name.ApiID {
			return fmt.Errorf("api mismatch in resource %s and target pattern %v", resourceName.String(), targetPattern)
		}
	default:
		return fmt.Errorf("unsupported resource type %T", targetPatternName)
	}

	// TODO: Filter check
	return nil
}

func CalculateScore(
	ctx context.Context,
	client connection.Client,
	definition *rpc.ScoreDefinition,
	resource patterns.ResourceInstance) error {

	project := fmt.Sprintf("%s/locations/global", resource.ResourceName().Project())

	// evaluate the expression and return a scoreValue
	scoreValue, err := processFormula(ctx, client, definition, resource)
	if err != nil {
		return err
	}

	// generate a score proto from the scoreValue
	score, err := processScoreType(definition, scoreValue, project)
	if err != nil {
		return err
	}

	// TODO: Add dry_run flag

	err = uploadScore(ctx, client, resource, score)
	if err != nil {
		return err
	}

	return nil

}

func processFormula(
	ctx context.Context,
	client connection.Client,
	definition *rpc.ScoreDefinition,
	resource patterns.ResourceInstance) (interface{}, error) {

	// Apply score formula
	switch formula := definition.GetFormula().(type) {
	case *rpc.ScoreDefinition_ScoreFormula:
		scoreValue, err := processScoreFormula(ctx, client, formula.ScoreFormula, resource)
		if err != nil {
			return nil, err
		}
		return scoreValue, nil
	// TODO: Add support for rollup_formula
	default:
		return nil, fmt.Errorf("invalid formula in ScoreDefinition: {%v} ", formula)
	}
}

func processScoreFormula(
	ctx context.Context,
	client connection.Client,
	formula *rpc.ScoreFormula,
	resource patterns.ResourceInstance) (interface{}, error) {

	// Fetch the artifact
	extendedArtifact, err := patterns.SubstituteReferenceEntity(formula.GetArtifact().GetPattern(), resource.ResourceName())
	if err != nil {
		return nil, fmt.Errorf("invalid score_formula.artifact.pattern: %s, %s", formula.GetArtifact().GetPattern(), err)
	}
	artifactName, err := names.ParseArtifact(extendedArtifact.String())
	if err != nil {
		return nil, fmt.Errorf("invalid score_formula.artifact.pattern: %s, %s", formula.GetArtifact().GetPattern(), err)
	}

	var contents []byte
	var mimeType string
	err = core.GetArtifact(ctx, client, artifactName, true, func(artifact *rpc.Artifact) error {
		contents = artifact.GetContents()
		mimeType = artifact.GetMimeType()
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artifact %s: %s", artifactName, err)
	}

	// TODO: Add timestamp check. Compute the score only if the artifact has an update since the last score calculation

	// Apply the score_expression
	value, err := evaluateScoreExpression(formula.GetScoreExpression(), mimeType, contents)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func processScoreType(definition *rpc.ScoreDefinition, scoreValue interface{}, project string) (*rpc.Score, error) {

	// Initialize Score proto
	score := &rpc.Score{
		Id:             fmt.Sprintf("score-%s", definition.GetId()),
		Kind:           "Score",
		DisplayName:    definition.GetDisplayName(),
		Description:    definition.GetDescription(),
		Uri:            definition.GetUri(),
		UriDisplayName: definition.GetUriDisplayName(),
		DefinitionName: fmt.Sprintf("%s/artifacts/%s", project, definition.GetId()),
	}

	// Set the Value field according to the type
	switch definition.GetType().(type) {
	case *rpc.ScoreDefinition_Integer:
		// Score proto expects int32 type
		value := int32(0)

		// Convert scoreValue to appropriate type
		// evaluateScoreExpression can return either a float or int value.
		// Both are valid for an integer.
		if intVal, ok := scoreValue.(int); ok {
			value = int32(intVal)
		} else if floatVal, ok := scoreValue.(float64); ok {
			value = int32(floatVal)
		} else {
			return nil, fmt.Errorf("failed typecheck for output: expected either int ot float64 got %s", scoreValue)
		}

		// Check that the scoreValue is within min/max limits
		configuredMin := definition.GetInteger().GetMinValue() // 0 if not set
		configuredMax := definition.GetInteger().GetMaxValue() // 0 if not set
		if value < configuredMin {
			return nil, fmt.Errorf("evaluated score value(%d) cannot be less than the configured min_value (%d)", value, configuredMin)
		}
		if value > configuredMax {
			return nil, fmt.Errorf("evaluated score value(%d) cannot be greater than the configured max_value (%d)", value, configuredMax)
		}

		// Populate Value field in Score proto
		score.Value = &rpc.Score_IntegerValue{
			IntegerValue: &rpc.IntegerValue{
				Value:    value,
				MinValue: configuredMin,
				MaxValue: configuredMax,
			},
		}

		// Populate the severity field according to Thresholds
		for _, t := range definition.GetInteger().GetThresholds() {
			if value >= t.GetRange().GetMin() && value <= t.GetRange().GetMax() {
				score.Severity = t.GetSeverity()
				break
			}
		}

	case *rpc.ScoreDefinition_Percent:
		// Score proto expects int32 type
		value := float32(0)

		// Convert scoreValue to appropriate type
		// evaluateScoreExpression can return either a float or int value.
		// Both are valid for an integer.
		if intVal, ok := scoreValue.(int); ok {
			value = float32(intVal)
		} else if floatVal, ok := scoreValue.(float64); ok {
			value = float32(floatVal)
		} else {
			return nil, fmt.Errorf("failed typecheck for output: expected either int ot float64 got %s", scoreValue)
		}

		// Check that the scoreValue is within min/max limits
		if value < 0 {
			return nil, fmt.Errorf("evaluated score value(%f) cannot be less than 0", value)
		}
		if value > 100 {
			return nil, fmt.Errorf("evaluated score value(%f) cannot be greater than 100", value)
		}

		// Populate Value field in Score proto
		score.Value = &rpc.Score_PercentValue{
			PercentValue: &rpc.PercentValue{
				Value: value,
			},
		}

		// Populate the severity field according to Thresholds
		for _, t := range definition.GetPercent().GetThresholds() {

			if value >= float32(t.GetRange().GetMin()) && value <= float32(t.GetRange().GetMax()) {
				score.Severity = t.GetSeverity()
				break
			}
		}

	case *rpc.ScoreDefinition_Boolean:
		// Convert scoreValue to appropriate type
		boolVal, ok := scoreValue.(bool)
		if !ok {
			return nil, fmt.Errorf("failed typecheck for output: expected bool")
		}

		displayValue := ""

		if boolVal {
			if configuredDisplay := definition.GetBoolean().GetDisplayTrue(); len(configuredDisplay) > 0 {
				displayValue = configuredDisplay
			} else {
				displayValue = "true"
			}
		} else {
			if configuredDisplay := definition.GetBoolean().GetDisplayFalse(); len(configuredDisplay) > 0 {
				displayValue = configuredDisplay
			} else {
				displayValue = "false"
			}
		}

		// Populate Value field in Score proto
		score.Value = &rpc.Score_BooleanValue{
			BooleanValue: &rpc.BooleanValue{
				Value:        boolVal,
				DisplayValue: displayValue,
			},
		}

		// Populate the severity field according to Thresholds
		for _, t := range definition.GetBoolean().GetThresholds() {
			if t.Value == boolVal {
				score.Severity = t.Severity
			}
		}
	}

	return score, nil
}

func uploadScore(ctx context.Context, client connection.Client, resource patterns.ResourceInstance, score *rpc.Score) error {

	artifactBytes, err := proto.Marshal(score)
	if err != nil {
		return err
	}
	artifact := &rpc.Artifact{
		Name:     fmt.Sprintf("%s/artifacts/%s", resource.ResourceName().String(), score.GetId()),
		Contents: artifactBytes,
		MimeType: patch.ScoreMimeType,
	}
	log.Debugf(ctx, "Uploading %s", artifact.GetName())
	if err = core.SetArtifact(ctx, client, artifact); err != nil {
		return fmt.Errorf("failed to save artifact %s: %s", artifact.GetName(), err)
	}

	return nil
}
