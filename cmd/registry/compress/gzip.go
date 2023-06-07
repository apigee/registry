// Copyright 2020 Google LLC.
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
	"compress/gzip"
	"io"
)

// GZippedBytes compresses a slice of bytes.
func GZippedBytes(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	_, err := zw.Write(input)
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GUnzippedBytes uncompresses a slice of bytes.
func GUnzippedBytes(input []byte) ([]byte, error) {
	buf := bytes.NewBuffer(input)
	zr, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(zr)
}
