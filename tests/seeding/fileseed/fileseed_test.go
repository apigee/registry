package fileseed

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestWrite(t *testing.T) {
	testDir := t.TempDir()
	files := []File{
		{
			Path:     filepath.Join(testDir, "foo/bar.txt"),
			Contents: []byte("foobar"),
		},
		{
			Path:     filepath.Join(testDir, "foo/baz.txt"),
			Contents: []byte("foobaz"),
		},
	}

	if err := Write(files...); err != nil {
		t.Errorf("Write(%+v) returned error: %s", files, err)
	}

	for _, want := range files {
		if contents, err := ioutil.ReadFile(want.Path); err != nil {
			t.Errorf("ReadFile(%q) returned error: %s", want.Path, err)
		} else if !bytes.Equal(contents, want.Contents) {
			t.Errorf("ReadFile(%q) got %q want %q", want.Path, contents, want.Contents)
		}
	}
}
