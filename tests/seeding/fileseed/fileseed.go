// Copyright 2021 Google LLC. All Rights Reserved.
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

package fileseed

import (
	"errors"
	"os"
	"path/filepath"
)

type File struct {
	Path     string
	Contents []byte
}

func Write(files ...File) error {
	var errs []error
	for _, f := range files {
		if err := os.MkdirAll(filepath.Dir(f.Path), os.ModePerm); err != nil {
			errs = append(errs, err)
			continue
		}

		fw, err := os.OpenFile(f.Path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		defer fw.Close()

		if _, err := fw.Write(f.Contents); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		msg := "failed to write one or more files:"
		for _, err := range errs {
			msg += "\n\t" + err.Error()
		}
		return errors.New(msg)
	}

	return nil
}
