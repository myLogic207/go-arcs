package collector

import (
	"git.mylogic.dev/homelab/go-arcs/pkg/store"
)

type Collector interface {
	store.Object
	Name() string
	GetHash() string
	SetHash(string)
}

type CollectorStore interface {
	store.Store[Collector]
}

type collector struct {
	id         string
	name       string
	attributes map[string]string
	hash       string
}

func New(id string, name string, attributes map[string]string, hash string) Collector {
	return &collector{
		id,
		name,
		attributes,
		hash,
	}
}

func (c *collector) ID() string {
	return c.id
}

func (c *collector) Name() string {
	return c.name
}

func (c *collector) Attributes() map[string]string {
	return c.attributes
}

func (c *collector) SetHash(cfg string) {
	c.hash = store.Hash([]byte(cfg))
}

func (c *collector) GetHash() string {
	return c.hash
}
