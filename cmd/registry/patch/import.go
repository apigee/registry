package patch

import "github.com/apigee/registry/rpc"

func (m *Manifest) Message() *rpc.Manifest {
	return &rpc.Manifest{
		Id:                 "",
		Kind:               "",
		GeneratedResources: m.generatedResources(),
	}
}

func (m *Manifest) generatedResources() []*rpc.GeneratedResource {
	v := make([]*rpc.GeneratedResource, 0)
	for _, g := range m.Body.GeneratedResources {
		v = append(v, &rpc.GeneratedResource{
			Pattern:      g.Pattern,
			Filter:       g.Filter,
			Receipt:      g.Receipt,
			Dependencies: m.dependencies(g),
			Action:       g.Action,
		})
	}
	return v
}

func (m *Manifest) dependencies(g *ManifestGeneratedResource) []*rpc.Dependency {
	v := make([]*rpc.Dependency, 0)
	for _, d := range g.Dependencies {
		v = append(v, &rpc.Dependency{
			Pattern: d.Pattern,
			Filter:  d.Filter,
		})
	}
	return v
}
