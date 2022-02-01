package patch

import (
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/protobuf/proto"
)

type UnknownArtifact struct {
	Header `yaml:",inline"`
	Body   struct {
		MimeType string `yaml:"mimeType,omitempty"`
	} `yaml:"body"`
}

func (a *UnknownArtifact) GetHeader() *Header {
	return &a.Header
}

func (a *UnknownArtifact) GetMessage() proto.Message {
	return nil
}

func newUnknownArtifact(message *rpc.Artifact) (*UnknownArtifact, error) {
	artifactName, err := names.ParseArtifact(message.Name)
	if err != nil {
		return nil, err
	}
	artifact := &UnknownArtifact{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "Artifact",
			Metadata: Metadata{
				Name: artifactName.ArtifactID(),
			},
		},
	}
	return artifact, nil
}
