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
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// UnzipArchiveToPath will decompress a zip archive, writing all files and folders
// within the zip archive (parameter 1) to an output directory (parameter 2).
// Based on an example published at https://golangcode.com/unzip-files-in-go/
func UnzipArchiveToPath(b []byte, dest string) ([]string, error) {
	var filenames []string
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return filenames, err
	}
	for _, f := range r.File {
		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)
		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}
		filenames = append(filenames, fpath)
		if f.FileInfo().IsDir() {
			// Make Folder
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}
		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		_, err = io.Copy(outFile, rc)
		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()
		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// UnzipArchiveToMap will decompress a zip archive to a map.
// May be memory intensive for large zip archives.
func UnzipArchiveToMap(b []byte) (map[string][]byte, error) {
	contents := make(map[string][]byte, 0)
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return contents, err
	}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return contents, err
		}
		bytes, err := io.ReadAll(rc)
		if err != nil {
			return contents, err
		}
		// Close the file without defer to close before next iteration of loop
		if err = rc.Close(); err != nil {
			return contents, err
		}
		contents[f.Name] = bytes
	}
	return contents, nil
}

// ZipArchiveOfPath reads the contents of a path into a zip archive.
// The specified prefix is stripped from file names in the archive.
// Based on an example published at https://golangcode.com/create-zip-files-in-go/
func ZipArchiveOfPath(path, prefix string, recursive bool) (buf bytes.Buffer, err error) {
	zipWriter := zip.NewWriter(&buf)
	defer zipWriter.Close()

	err = filepath.WalkDir(path, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if entry.IsDir() && p != path && !recursive {
			return filepath.SkipDir // Skip the directory and contents.
		} else if entry.IsDir() {
			return nil // Do nothing for the directory, but still walk its contents.
		}
		if err = addFileToZip(zipWriter, p, prefix); err != nil {
			return err
		}
		return nil
	})
	return buf, err
}

// ZipArchiveOfFiles stores a list of files in a zip archive.
// The specified prefix is stripped from file names in the archive.
func ZipArchiveOfFiles(files []string, prefix string) (buf bytes.Buffer, err error) {
	zipWriter := zip.NewWriter(&buf)
	defer zipWriter.Close()

	for _, filename := range files {
		if err = addFileToZip(zipWriter, prefix+filename, prefix); err != nil {
			return buf, err
		}
	}
	return buf, err
}

func addFileToZip(zipWriter *zip.Writer, filename, prefix string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()
	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	name := strings.TrimPrefix(filename, prefix)
	header.Name = name
	// Set to Deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
