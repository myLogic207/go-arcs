package config

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
)

var (
	ErrFileOpen = errors.New("failed to open file")
	ErrFileRead = errors.New("failed to read file")
)

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

	return content, nil
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
