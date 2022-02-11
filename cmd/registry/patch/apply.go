package patch

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"gopkg.in/yaml.v2"
)

func ApplyDirectory(
	ctx context.Context,
	client connection.Client,
	fileName string,
	parent string) error {
	return filepath.Walk(fileName,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			return ApplyFile(ctx, client, path, parent)
		})
}

func ApplyFile(
	ctx context.Context,
	client connection.Client,
	fileName string,
	parent string) error {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to read file")
	}

	// get the id and kind of artifact from the YAML elements common to all artifacts
	var header Header
	err = yaml.Unmarshal(bytes, &header)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to parse YAML")
	}
	if header.APIVersion != REGISTRY_V1 {
		log.FromContext(ctx).Fatalf("Unsupported API version: %s", header.APIVersion)
	}
	switch header.Kind {
	case "API":
		var api API
		err = yaml.Unmarshal(bytes, &api)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to parse YAML")
		}
		err = applyApiPatch(ctx, client, &api, parent)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to apply patch")
		}
	case "Lifecycle":
		var lifecycle Lifecycle
		err = yaml.Unmarshal(bytes, &lifecycle)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to parse YAML")
		}
		err = applyArtifactPatch(ctx, client, &lifecycle, parent)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to apply patch")
		}
	case "Manifest":
		var manifest Manifest
		err = yaml.Unmarshal(bytes, &manifest)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to parse YAML")
		}
		err = applyArtifactPatch(ctx, client, &manifest, parent)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to apply patch")
		}
	case "TaxonomyList":
		var taxonomyList TaxonomyList
		err = yaml.Unmarshal(bytes, &taxonomyList)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to parse YAML")
		}
		err = applyArtifactPatch(ctx, client, &taxonomyList, parent)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to apply patch")
		}
	default:
		log.FromContext(ctx).Fatalf("Unsupported kind: %s", header.Kind)
	}
	return nil
}
