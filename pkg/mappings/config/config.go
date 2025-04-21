package config

import (
	"context"
	"errors"
	"strings"

	"git.mylogic.dev/homelab/go-arcs/pkg/store"
)

type Protocol string

const (
	ProtoDelimiter = "://"
	FileProto      = Protocol("file")
	FTPProto       = Protocol("ftp")
	HTTPProto      = Protocol("http")
)

var (
	ErrProtoUnknown = errors.New("could not identify protocol")
	ErrProtoParts   = errors.New("source malformed, make sure source looks like [proto]://[source]")
)

type ContentHandler func(context.Context, string) (string, error)

type Config interface {
	store.Object
	Content(context.Context) (string, error)
	Source() string
}

type ConfigStore interface {
	store.Store[Config]
}

type config struct {
	id             string
	protocol       Protocol
	path           string
	attributes     map[string]string
	contentHandler ContentHandler
}

func checkProto(proto string) (Protocol, error) {
	switch proto {
	case "file":
		return FileProto, nil
	case "ftp":
		return FTPProto, nil
	case "http":
	case "https":
		return HTTPProto, nil
	}
	return "", ErrProtoUnknown
}

func getContentHandler(proto Protocol) ContentHandler {
	switch proto {
	case FileProto:
		return fileHandler
	case FTPProto:
		return ftpHandler
	case HTTPProto:
		return httpHandler
	}
	return nil
}

func New(source string, attributes map[string]string) (Config, error) {
	id := store.Hash([]byte(source))
	protocolRaw, path, found := strings.Cut(source, ProtoDelimiter)
	if !found {
		return nil, ErrProtoParts
	}
	protocol, err := checkProto(protocolRaw)
	if err != nil {
		return nil, err
	}
	contentHandler := getContentHandler(protocol)

	return &config{
		id,
		protocol,
		path,
		attributes,
		contentHandler,
	}, nil
}

func (c *config) ID() string {
	return c.id
}

func (c *config) Attributes() map[string]string {
	// check if any filter is provided for quick return
	return c.attributes
}

func (c *config) Content(ctx context.Context) (string, error) {
	var handler func(context.Context, string) (string, error)
	switch c.protocol {
	case FileProto:
		handler = fileHandler
	case FTPProto:
		handler = ftpHandler
	case HTTPProto:
		handler = httpHandler
	}
	if nil == handler {
		return "", errors.New("")
	}
	return handler(ctx, c.path)
}

func fileHandler(ctx context.Context, path string) (string, error) {
	return "", nil
}

func ftpHandler(ctx context.Context, path string) (string, error) {
	return "", nil
}

func httpHandler(ctx context.Context, path string) (string, error) {
	return "", nil
}

func (c *config) Source() string {
	return strings.Join([]string{string(c.protocol), c.path}, ProtoDelimiter)
}
