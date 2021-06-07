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
