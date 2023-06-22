// Copyright 2022 Google LLC.
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

package buildinfo

import (
	"runtime/debug"

	"github.com/apigee/registry/rpc"
)

func module(m *debug.Module) *rpc.BuildInfo_Module {
	if m == nil {
		return nil
	}
	return &rpc.BuildInfo_Module{
		Path:        m.Path,
		Version:     m.Version,
		Sum:         m.Sum,
		Replacement: module(m.Replace),
	}
}

func BuildInfo() *rpc.BuildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	settings := make(map[string]string, 0)
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	dependencies := make([]*rpc.BuildInfo_Module, 0)
	for _, dep := range info.Deps {
		dependencies = append(dependencies, module(dep))
	}
	return &rpc.BuildInfo{
		GoVersion:    info.GoVersion,
		Path:         info.Path,
		Main:         module(&info.Main),
		Dependencies: dependencies,
		Settings:     settings,
	}
}
