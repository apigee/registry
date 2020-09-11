package tools

import (
	"github.com/apigee/registry/rpc"
)

type ProjectHandler func(*rpc.Project)
type ApiHandler func(*rpc.Api)
type VersionHandler func(*rpc.Version)
type SpecHandler func(*rpc.Spec)
type PropertyHandler func(*rpc.Property)
type LabelHandler func(*rpc.Label)
