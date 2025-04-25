package config

import (
	"context"
	"os"
	"testing"

	"git.mylogic.dev/homelab/go-arcs/pkg/store"
	"github.com/stretchr/testify/assert"
)

func Test_ParseRaw(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []Config
		wantErr bool
	}{
		{
			name: "Load Simple File",
			args: args{
				content: []byte(`
- source: 'file://test'
  attributes:
    test: value`),
			},
			want: []Config{
				&config{
					protocol: "file",
					path:     "test",
					attributes: map[string]string{
						"test": "value",
					},
					id: store.Hash([]byte("file://test")),
				},
			},
			wantErr: false,
		},
		{
			name: "Load list file",
			args: args{
				content: []byte(`
- source: 'file://test1'
  attributes:
    test: value
- source: 'file://test2'
  attributes:
    test: value
    test2: value2`),
			},
			want: []Config{
				&config{
					protocol: "file",
					path:     "test1",
					attributes: map[string]string{
						"test": "value",
					},
					id: store.Hash([]byte("file://test1")),
				},
				&config{
					protocol: "file",
					path:     "test2",
					attributes: map[string]string{
						"test":  "value",
						"test2": "value2",
					},
					id: store.Hash([]byte("file://test2")),
				},
			},
		},
		{
			name: "Fail malformed path",
			args: args{
				content: []byte(`
- source: 'test/path'
  attributes:
    test: value`),
			},
			want:    []Config{nil},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConfig(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !assert.Equal(t, tt.want, got) {
				t.Errorf("loadFile() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_LoadFile(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []Config
		wantErr bool
	}{
		{
			name: "simple file",
			args: args{
				content: []byte(`
- source: 'file://test'
  attributes:
    test: value`),
			},
			want: []Config{
				&config{
					protocol: "file",
					path:     "test",
					attributes: map[string]string{
						"test": "value",
					},
					id: store.Hash([]byte("file://test")),
				},
			},
			wantErr: false,
		},
	}
	testDir, err := os.MkdirTemp(".", "test_directory")
	defer os.RemoveAll(testDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp(testDir, "test_file")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := file.Write(tt.args.content); err != nil {
				t.Fatal(err)
			}

			got, err := Load(context.Background(), file.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("loadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.want, got) {
				t.Errorf("loadFile() = %#v, want %#v", got, tt.want)
			}
		})
	}

}
