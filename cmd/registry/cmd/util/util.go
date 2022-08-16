// Copyright 2022 Google LLC. All Rights Reserved.
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

package util

import (
	"fmt"
	"path"
	"strings"

	"github.com/apigee/registry/pkg/config"
	"github.com/apigee/registry/pkg/connection"
)

var NoActiveConfigurationError = fmt.Errorf(`No active configuration. Use 'registry config configurations' to manage.`)

func TargetConfiguration() (name string, c config.Configuration, err error) {
	name, c, err = config.ActiveRaw()
	if name == "" {
		return name, config.Configuration{}, NoActiveConfigurationError
	}

	return
}

// FQName ensures the project and location, if available,
// are properly included in the resource name.
func FQName(c connection.Config, name string) string {
	name = strings.TrimPrefix(name, "/")
	if !strings.HasPrefix(name, "projects") && c.Project != "" {
		if strings.HasPrefix(name, "locations") {
			name = path.Join("projects", c.Project, name)
		} else if c.Location != "" {
			name = path.Join("projects", c.Project, "locations", c.Location, name)
		}
	}
	return name
}
