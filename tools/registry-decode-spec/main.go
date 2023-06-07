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

package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/hex"
	"io"
	"log"
	"os"
)

// This is equivalent to running
//
//	base64 --decode | gunzip
//
// to decode specs returned by the `registry rpc` subcommands, with an additional effort to
// handle the hex-encoded inputs produced by registry-encode-spec.
func main() {
	// decode the spec
	reader := bufio.NewReader(os.Stdin)
	s, _ := reader.ReadString('\n')
	decoded, err := hex.DecodeString(s)
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(s)
	}
	if err != nil {
		log.Fatal(err)
	}

	// gunzip the spec
	buf := bytes.NewBuffer(decoded)
	zr, err := gzip.NewReader(buf)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := io.Copy(os.Stdout, zr); err != nil {
		log.Fatal(err)
	}
	if err := zr.Close(); err != nil {
		log.Fatal(err)
	}
}
