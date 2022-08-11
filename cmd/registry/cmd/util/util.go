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

	"github.com/apigee/registry/pkg/config"
)

var NoActiveConfigurationError = fmt.Errorf(`No active configuration. Use 'registry config configurations' to manage.`)

func TargetConfiguration() (name string, c config.Configuration, err error) {
	name, c, err = config.ActiveRaw()
	if name == "" {
		return name, config.Configuration{}, NoActiveConfigurationError
	}

	return
}
