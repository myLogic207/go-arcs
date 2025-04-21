package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetConfig(t *testing.T) {
	ctx := context.Background()
	id := "foobar"
	attributes := map[string]string{
		"test": "value",
	}
	store := NewStore[Object](nil, nil)
	storedID, err := store.Set(ctx, &object{attributes: attributes, id: id})

	assert.NoError(t, err)
	assert.Equal(t, store.Get(ctx, storedID).ID(), storedID)
	assert.Len(t, store.GetByAttributes(ctx, attributes), 1)
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
