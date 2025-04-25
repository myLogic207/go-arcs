package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_loadDir(t *testing.T) {
	testDir, err := os.MkdirTemp(".", "test_directory")
	defer os.RemoveAll(testDir)
	if err != nil {
		t.Fatal(err)
	}
	wantFiles := make([]string, 10)
	for i := range len(wantFiles) {
		file, err := os.CreateTemp(testDir, "test_file")
		if err != nil {
			t.Fatal(err)
		}
		wantFiles[i], err = filepath.Abs(file.Name())
		if err != nil {
			t.Fatal(err)
		}
	}

	foundFiles, err := GetFileOrFiles(testDir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, foundFiles, len(wantFiles))
	assert.ElementsMatch(t, foundFiles, wantFiles)
}
