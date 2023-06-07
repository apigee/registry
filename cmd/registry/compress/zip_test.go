// Copyright 2023 Google LLC.
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

package compress

import (
	"bytes"
	"testing"
)

func TestZipArchiveOfPath(t *testing.T) {
	tests := []struct {
		name      string
		recursive bool
		expected  map[string][]byte
	}{
		{
			"recursive",
			true,
			map[string][]byte{
				"one":             []byte("one\n"),
				"two":             []byte("two\n"),
				"three/three":     []byte("three\n"),
				"three/four/four": []byte("four\n"),
				"five":            []byte("five\n"),
			},
		},
		{
			"nonrecursive",
			false,
			map[string][]byte{
				"one":  []byte("one\n"),
				"two":  []byte("two\n"),
				"five": []byte("five\n"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := "testdata/sample"
			b, err := ZipArchiveOfPath(path, "testdata/sample/", test.recursive)
			if err != nil {
				t.Fatalf("Failed to zip path %s", path)
			}
			a := b.Bytes()
			m, err := UnzipArchiveToMap(a)
			if err != nil {
				t.Fatalf("Failed to unzip archive %s to map", path)
			}
			for k, v := range test.expected {
				if !bytes.Equal(m[k], v) {
					t.Errorf("failed to get file %s from archive", k)
				}
			}
			for k, v := range m {
				if !bytes.Equal(v, test.expected[k]) {
					t.Errorf("found extra file %s in archive", k)
				}
			}
			output := t.TempDir()
			files, err := UnzipArchiveToPath(a, output)
			if err != nil {
				t.Fatalf("Failed to unzip archive %s to directory", path)
			}
			if len(files) != len(test.expected) {
				t.Errorf("unzipped incorrect number of files %d, expected %d", len(files), len(test.expected))
			}
		})
	}
}

func TestZipArchiveOfFiles(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected map[string][]byte
	}{
		{
			"files-123",
			[]string{"one", "two", "three/three"},
			map[string][]byte{
				"one":         []byte("one\n"),
				"two":         []byte("two\n"),
				"three/three": []byte("three\n"),
			},
		},
		{
			"files-45",
			[]string{"three/four/four", "five"},
			map[string][]byte{
				"three/four/four": []byte("four\n"),
				"five":            []byte("five\n"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := "testdata/sample"
			b, err := ZipArchiveOfFiles(test.files, "testdata/sample/")
			if err != nil {
				t.Fatalf("Failed to zip path %s", path)
			}
			m, err := UnzipArchiveToMap(b.Bytes())
			if err != nil {
				t.Fatalf("Failed to unzip archive %s", path)
			}
			for k, v := range test.expected {
				if !bytes.Equal(m[k], v) {
					t.Errorf("failed to get file %s from archive", k)
				}
			}
			for k, v := range m {
				if !bytes.Equal(v, test.expected[k]) {
					t.Errorf("found extra file %s in archive", k)
				}
			}
		})
	}
}
