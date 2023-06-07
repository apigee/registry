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

package connection

import (
	"testing"
)

func TestFQNamePartialConfig(t *testing.T) {
	data := []struct {
		input string
		want  string
	}{
		{"projects/project1", "projects/project1"},
		{"/projects/project1", "projects/project1"},
		{"locations/foo", "locations/foo"},
		{"/locations/foo", "locations/foo"},
		{"apis/foo", "apis/foo"},
		{"/apis/foo", "apis/foo"},
	}

	c := Config{
		Location: "location",
	}
	for _, d := range data {
		t.Run(d.input, func(t *testing.T) {
			got := c.FQName(d.input)
			if d.want != got {
				t.Errorf("want: %q, got: %q", d.want, got)
			}
		})
	}
}

func TestFQNameUnspecifiedLocation(t *testing.T) {
	data := []struct {
		input string
		want  string
	}{
		{"locations/foo", "projects/project1/locations/foo"},
		{"/locations/foo", "projects/project1/locations/foo"},
		{"apis/foo", "projects/project1/locations/global/apis/foo"},
		{"/apis/foo", "projects/project1/locations/global/apis/foo"},
	}

	c := Config{
		Project: "project1",
	}
	for _, d := range data {
		t.Run(d.input, func(t *testing.T) {
			got := c.FQName(d.input)
			if d.want != got {
				t.Errorf("want: %q, got: %q", d.want, got)
			}
		})
	}
}

func TestFQNameFullConfig(t *testing.T) {
	data := []struct {
		input string
		want  string
	}{
		{"projects/foo", "projects/foo"},
		{"/projects/foo", "projects/foo"},
		{"", "projects/project1/locations/location1"},
		{"/", "projects/project1/locations/location1"},
		{"locations/foo", "projects/project1/locations/foo"},
		{"/locations/foo", "projects/project1/locations/foo"},
		{"projects/foo/locations/bar", "projects/foo/locations/bar"},
		{"/apis/foo", "projects/project1/locations/location1/apis/foo"},
	}

	c := Config{
		Project:  "project1",
		Location: "location1",
	}
	for _, d := range data {
		got := c.FQName(d.input)
		if d.want != got {
			t.Errorf("for: %q, want: %q, got: %q", d.input, d.want, got)
		}
	}
}
