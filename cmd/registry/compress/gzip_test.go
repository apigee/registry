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

import "testing"

func TestGZip(t *testing.T) {
	initial := []byte("hello")
	compressed, err := GZippedBytes(initial)
	if err != nil {
		t.Fatal("Failed to compress message")
	}
	final, err := GUnzippedBytes(compressed)
	if err != nil {
		t.Fatal("Failed to uncompress message")
	}
	if string(final) != string(initial) {
		t.Error("Failed to preserve message through compression and uncompression")
	}
	_, err = GUnzippedBytes(initial)
	if err == nil {
		t.Error("Uncompression of uncompressed data succeeded and should have failed")
	}
}
