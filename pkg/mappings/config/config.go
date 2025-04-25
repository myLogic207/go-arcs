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
	ErrGetContent   = errors.New("could not get config content")
)

type ContentHandler func(context.Context, string) ([]byte, error)

type Config interface {
	store.Object
	Content(context.Context) (string, error)
	Source() string
}

type Store interface {
	store.Store[Config]
}

type config struct {
	id         string
	protocol   Protocol
	path       string
	attributes map[string]string
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

	return &config{
		id,
		protocol,
		path,
		attributes,
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
	contentHandler := getContentHandler(c.protocol)
	content, err := contentHandler(ctx, c.path)
	if err != nil {
		return "", errors.Join(ErrGetContent, err)
	}
	return string(content), nil
}

func (c *config) Source() string {
	return strings.Join([]string{string(c.protocol), c.path}, ProtoDelimiter)
}
