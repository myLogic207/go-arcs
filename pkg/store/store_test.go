package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetConfig(t *testing.T) {
	type args struct {
		id         string
		attributes map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Simple Load",
			args: args{
				id: "foobar",
				attributes: map[string]string{
					"test": "value",
				},
			},
			want:    "foobar",
			wantErr: false,
		},
	}

	ctx := context.Background()
	store := NewStore[Object](nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Set(ctx, &object{attributes: tt.args.attributes, id: tt.args.id})
			if (err != nil) != tt.wantErr {
				t.Errorf("store.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !assert.Equal(t, got, tt.want) {
				t.Errorf("store.Set() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	type args struct {
		objects []Object
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "load one",
			args: args{
				objects: []Object{
					&object{
						id: "foobar",
						attributes: map[string]string{
							"test": "value",
						},
					},
				},
			},
			want:    []string{"foobar"},
			wantErr: false,
		},
		{
			name: "load many",
			args: args{
				objects: []Object{
					&object{
						id: "foobar1",
						attributes: map[string]string{
							"test": "value",
						},
					},
					&object{
						id: "foobar2",
						attributes: map[string]string{
							"test":  "value",
							"test2": "value2",
						},
					},
				},
			},
			want:    []string{"foobar1", "foobar2"},
			wantErr: false,
		},
		{
			name: "overwrite one",
			args: args{
				objects: []Object{
					&object{
						id: "foobar",
						attributes: map[string]string{
							"test": "value",
						},
					},
					&object{
						id: "foobar",
						attributes: map[string]string{
							"test": "value",
						},
					},
				},
			},
			want:    []string{"foobar", "foobar"},
			wantErr: false,
		},
	}

	ctx := context.Background()
	store := NewStore[Object](nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Load(ctx, tt.args.objects)
			if (err != nil) != tt.wantErr {
				t.Errorf("store.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !assert.Equal(t, got, tt.want) {
				t.Errorf("store.Set() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	ctx := context.Background()
	id := "test"
	attributes := map[string]string{
		"test": "attribute",
	}
	store := NewStore[Object](
		map[string]Object{
			"hash":   &object{id: id, attributes: attributes},
			"noHash": &object{id: "NoAssert", attributes: nil},
		},
		map[string]map[string]map[string]bool{
			"test": {
				"attribute": {
					"hash":   true,
					"unHash": false,
				},
				"foo": {
					"NoHash": true,
				},
			},
			"tested": {
				"string": {
					"NoHash": false,
				},
			},
		})

	configs := store.GetByAttributes(ctx, attributes)
	assert.Len(t, configs, 1)
	assert.Equal(t, configs[0].ID(), id)
}

func TestRemoveConfig(t *testing.T) {
	ctx := context.Background()
	id := "test"
	attributes := map[string]string{
		"test": "value",
	}
	storeID := "hash"
	store := NewStore[Object](
		map[string]Object{
			storeID:  &object{id: id, attributes: attributes},
			"noHash": &object{id: "NoAssert", attributes: nil},
		},
		map[string]map[string]map[string]bool{
			"test": {
				"value": {
					"hash":   true,
					"unHash": false,
				},
				"foo": {
					"NoHash": true,
				},
			},
			"tested": {
				"string": {
					"NoHash": false,
				},
			},
		},
	)

	assert.Equal(t, store.Get(ctx, storeID).ID(), id)
	store.Remove(ctx, storeID)
	assert.Nil(t, store.Get(ctx, storeID))
}
