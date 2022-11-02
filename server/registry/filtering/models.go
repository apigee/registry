package filtering

import (
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/apigee/registry/server/registry/names"
)

var ProjectFields = map[string]FieldType{
	"name":         String,
	"project_id":   String,
	"display_name": String,
	"description":  String,
	"create_time":  Timestamp,
	"update_time":  Timestamp,
}

func ProjectMap(p models.Project) (map[string]interface{}, error) {
	return map[string]interface{}{
		"name":         p.Name(),
		"project_id":   p.ProjectID,
		"display_name": p.DisplayName,
		"description":  p.Description,
		"create_time":  p.CreateTime,
		"update_time":  p.UpdateTime,
	}, nil
}

func ProjectMapFromMessage(projectMessage *rpc.Project) (map[string]interface{}, error) {
	projectName, err := names.ParseProjectWithLocation(projectMessage.GetName())
	if err != nil {
		projectName, err = names.ParseProject(projectMessage.GetName())
	}
	if err != nil {
		return nil, err
	}
	project := models.NewProject(projectName, projectMessage)
	return ProjectMap(*project)

}

var ApiFields = map[string]FieldType{
	"name":                   String,
	"project_id":             String,
	"api_id":                 String,
	"display_name":           String,
	"description":            String,
	"create_time":            Timestamp,
	"update_time":            Timestamp,
	"availability":           String,
	"recommended_version":    String,
	"recommended_deployment": String,
	"labels":                 StringMap,
}

func ApiMap(api models.Api) (map[string]interface{}, error) {
	labels, err := api.LabelsMap()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":                api.Name(),
		"project_id":          api.ProjectID,
		"api_id":              api.ApiID,
		"display_name":        api.DisplayName,
		"description":         api.Description,
		"create_time":         api.CreateTime,
		"update_time":         api.UpdateTime,
		"availability":        api.Availability,
		"recommended_version": api.RecommendedVersion,
		"labels":              labels,
	}, nil
}

func ApiMapFromMessage(apiMessage *rpc.Api) (map[string]interface{}, error) {
	apiName, err := names.ParseApi(apiMessage.GetName())
	if err != nil {
		return nil, err
	}
	api, err := models.NewApi(apiName, apiMessage)
	if err != nil {
		return nil, err
	}
	return ApiMap(*api)
}

var VersionFields = map[string]FieldType{
	"name":         String,
	"project_id":   String,
	"api_id":       String,
	"version_id":   String,
	"display_name": String,
	"description":  String,
	"create_time":  Timestamp,
	"update_time":  Timestamp,
	"state":        String,
	"labels":       StringMap,
}

func VersionMap(version models.Version) (map[string]interface{}, error) {
	labels, err := version.LabelsMap()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":         version.Name(),
		"project_id":   version.ProjectID,
		"version_id":   version.VersionID,
		"display_name": version.DisplayName,
		"description":  version.Description,
		"create_time":  version.CreateTime,
		"update_time":  version.UpdateTime,
		"state":        version.State,
		"labels":       labels,
	}, nil
}

func VersionMapFromMessage(versionMessage *rpc.ApiVersion) (map[string]interface{}, error) {
	versionName, err := names.ParseVersion(versionMessage.GetName())
	if err != nil {
		return nil, err
	}
	version, err := models.NewVersion(versionName, versionMessage)
	if err != nil {
		return nil, err
	}
	return VersionMap(*version)
}

var SpecFields = map[string]FieldType{
	"name":                 String,
	"project_id":           String,
	"api_id":               String,
	"version_id":           String,
	"spec_id":              String,
	"filename":             String,
	"description":          String,
	"create_time":          Timestamp,
	"revision_create_time": Timestamp,
	"revision_update_time": Timestamp,
	"mime_type":            String,
	"size_bytes":           Int,
	"source_uri":           String,
	"labels":               StringMap,
}

func SpecMap(spec models.Spec) (map[string]interface{}, error) {
	labels, err := spec.LabelsMap()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":                 spec.Name(),
		"project_id":           spec.ProjectID,
		"api_id":               spec.ApiID,
		"version_id":           spec.VersionID,
		"spec_id":              spec.SpecID,
		"filename":             spec.FileName,
		"description":          spec.Description,
		"revision_id":          spec.RevisionID,
		"create_time":          spec.CreateTime,
		"revision_create_time": spec.RevisionCreateTime,
		"revision_update_time": spec.RevisionUpdateTime,
		"mime_type":            spec.MimeType,
		"size_bytes":           spec.SizeInBytes,
		"hash":                 spec.Hash,
		"source_uri":           spec.SourceURI,
		"labels":               labels,
	}, nil
}

func SpecMapFromMessage(specMessage *rpc.ApiSpec) (map[string]interface{}, error) {
	specName, err := names.ParseSpec(specMessage.GetName())
	if err != nil {
		return nil, err
	}
	spec, err := models.NewSpec(specName, specMessage)
	if err != nil {
		return nil, err
	}
	return SpecMap(*spec)
}

var DeploymentFields = map[string]FieldType{
	"name":                 String,
	"project_id":           String,
	"api_id":               String,
	"deployment_id":        String,
	"display_name":         String,
	"description":          String,
	"create_time":          Timestamp,
	"revision_create_time": Timestamp,
	"revision_update_time": Timestamp,
	"api_spec_revision":    String,
	"endpoint_uri":         String,
	"external_channel_uri": String,
	"intended_audience":    String,
	"access_guidance":      String,
	"labels":               StringMap,
}

func DeploymentMap(deployment models.Deployment) (map[string]interface{}, error) {
	labels, err := deployment.LabelsMap()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":                 deployment.Name(),
		"project_id":           deployment.ProjectID,
		"api_id":               deployment.ApiID,
		"deployment_id":        deployment.DeploymentID,
		"revision_id":          deployment.RevisionID,
		"display_name":         deployment.DisplayName,
		"description":          deployment.Description,
		"create_time":          deployment.CreateTime,
		"revision_create_time": deployment.RevisionCreateTime,
		"revision_update_time": deployment.RevisionUpdateTime,
		"api_spec_revision":    deployment.ApiSpecRevision,
		"endpoint_uri":         deployment.EndpointURI,
		"external_channel_uri": deployment.ExternalChannelURI,
		"intended_audience":    deployment.IntendedAudience,
		"access_guidance":      deployment.AccessGuidance,
		"labels":               labels,
	}, nil
}

func DeploymentMapFromMessage(deploymentMessage *rpc.ApiDeployment) (map[string]interface{}, error) {
	deploymentName, err := names.ParseDeployment(deploymentMessage.GetName())
	if err != nil {
		return nil, err
	}
	deployment, err := models.NewDeployment(deploymentName, deploymentMessage)
	if err != nil {
		return nil, err
	}
	return DeploymentMap(*deployment)
}

var ArtifactFields = map[string]FieldType{
	"name":        String,
	"project_id":  String,
	"api_id":      String,
	"version_id":  String,
	"spec_id":     String,
	"artifact_id": String,
	"create_time": Timestamp,
	"update_time": Timestamp,
	"mime_type":   String,
	"size_bytes":  Int,
}

func ArtifactMap(artifact models.Artifact) (map[string]interface{}, error) {
	return map[string]interface{}{
		"name":        artifact.Name(),
		"project_id":  artifact.ProjectID,
		"api_id":      artifact.ApiID,
		"version_id":  artifact.VersionID,
		"spec_id":     artifact.SpecID,
		"artifact_id": artifact.ArtifactID,
		"create_time": artifact.CreateTime,
		"update_time": artifact.UpdateTime,
		"mime_type":   artifact.MimeType,
		"size_bytes":  artifact.SizeInBytes,
	}, nil
}

func ArtifactMapFromMessage(artifactMessage *rpc.Artifact) (map[string]interface{}, error) {
	artifactName, err := names.ParseArtifact(artifactMessage.GetName())
	if err != nil {
		return nil, err
	}
	artifact, err := models.NewArtifact(artifactName, artifactMessage)
	if err != nil {
		return nil, err
	}
	return ArtifactMap(*artifact)
}
