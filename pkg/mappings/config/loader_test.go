package config

import (
	"testing"

	"git.mylogic.dev/homelab/go-arcs/pkg/store"
	"github.com/stretchr/testify/assert"
)

func Test_loadFile(t *testing.T) {
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
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRawConfig(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !assert.Equal(t, got, tt.want) {
				t.Errorf("loadFile() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
