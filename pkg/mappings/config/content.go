package config

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrDirOpen  = errors.New("failed to open directory")
	ErrFileOpen = errors.New("failed to open file")
	ErrFileRead = errors.New("failed to read file")
)

// check if the given path is a file or directory,
// always returns a slice of absolute paths
func GetFileOrFiles(name string) ([]string, error) {
	abs, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return []string{abs}, nil
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		names = append(names, filepath.Join(abs, ent.Name()))
	}

	return names, nil
}

func fileHandler(ctx context.Context, path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Join(ErrFileOpen, err)
	}
	defer file.Close()
	content, err := ioRead(file)
	if err != nil {
		return nil, errors.Join(ErrFileRead, err)
	}

	return content, ctx.Err()
}

func ftpHandler(ctx context.Context, path string) ([]byte, error) {
	return nil, nil
}

func httpHandler(ctx context.Context, path string) ([]byte, error) {
	return nil, nil
}

func ioRead(input io.Reader) ([]byte, error) {
	buf := make([]byte, 1024)
	var output bytes.Buffer

	for {
		n, err := input.Read(buf)

		if err == io.EOF {
			break // End of file, break the loop
		}
		if err != nil {
			return nil, err
		}
		output.Write(buf[:n])
	}
	return output.Bytes(), nil
}
