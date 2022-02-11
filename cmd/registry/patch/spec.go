package patch

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

type APISpec struct {
	Header `yaml:",inline"`
	Data   struct {
		FileName    string      `yaml:"filename,omitempty"`
		Description string      `yaml:"description,omitempty"`
		MimeType    string      `yaml:"mimeType,omitempty"`
		SourceURI   string      `yaml:"sourceURI,omitempty"`
		Artifacts   []*Artifact `yaml:"artifacts,omitempty"`
	} `yaml:"data"`
}

func newAPISpec(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiSpec) (*APISpec, error) {
	specName, err := names.ParseSpec(message.Name)
	if err != nil {
		return nil, err
	}
	spec := &APISpec{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "APISpec",
			Metadata: Metadata{
				Name:        specName.SpecID,
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
	}
	spec.Data.FileName = message.Filename
	spec.Data.Description = message.Description
	spec.Data.MimeType = message.MimeType
	spec.Data.SourceURI = message.SourceUri
	return spec, nil
}

func applyApiSpecPatch(
	ctx context.Context,
	client connection.Client,
	spec *APISpec,
	parent string) error {
	name := fmt.Sprintf("%s/specs/%s", parent, spec.Metadata.Name)
	req := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:        name,
			Filename:    spec.Data.FileName,
			Description: spec.Data.Description,
			MimeType:    spec.Data.MimeType,
			SourceUri:   spec.Data.SourceURI,
			Labels:      spec.Metadata.Labels,
			Annotations: spec.Metadata.Annotations,
		},
		AllowMissing: true,
	}
	// TODO: add support for multi-file specs
	// TODO: add support for local file import (maybe?)
	// TODO: verify mime type
	if spec.Data.SourceURI != "" {
		resp, err := http.Get(spec.Data.SourceURI)
		if err != nil {
			return err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if strings.Contains(spec.Data.MimeType, "+gzip") {
			body, err = core.GZippedBytes(body)
			if err != nil {
				return err
			}
		}
		req.ApiSpec.MimeType = spec.Data.MimeType
		req.ApiSpec.Contents = body
	}
	_, err := client.UpdateApiSpec(ctx, req)
	return err
}
