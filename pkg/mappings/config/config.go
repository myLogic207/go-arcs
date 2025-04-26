package config

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"git.mylogic.dev/homelab/go-arcs/pkg/store"
)

const ProtoDelimiter = "://"

var (
	ErrProtoUnknown = errors.New("could not identify protocol")
	ErrProtoParts   = errors.New("source malformed, make sure source looks like [proto]://[source]")
	ErrGetContent   = errors.New("could not get config content")

	knownProtocols = []string{
		"file",
		"http",
		"https",
	}
)

type ContentHandler func(context.Context, string, ...any) ([]byte, error)

type Config interface {
	store.Object
	Content(context.Context, ...any) (string, error)
	Source() string
}

type Store interface {
	store.Store[Config]
}

type config struct {
	id         string
	protocol   string
	path       string
	attributes map[string]string
}

func New(source string, attributes map[string]string) (Config, error) {
	id := store.Hash([]byte(source))
	protocol, path, found := strings.Cut(source, ProtoDelimiter)
	if !found {
		return nil, ErrProtoParts
	}

	if !slices.Contains(knownProtocols, protocol) {
		return nil, ErrProtoUnknown
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

func (c *config) Content(ctx context.Context, options ...any) (string, error) {
	var contentHandler func(context.Context, string) ([]byte, error)
	switch c.protocol {
	case "file":
		contentHandler = fileHandler
	// case FTPProto:
	// 	return ftpHandler
	case "https":
		fallthrough
	case "http":
		contentHandler = func(ctx context.Context, s string) ([]byte, error) {
			url := fmt.Sprintf("%v%v%v", c.protocol, ProtoDelimiter, c.path)
			headers := options[0].(http.Header)
			return httpHandler(ctx, url, headers)
		}
	default:
		return "", ErrProtoUnknown
	}
	content, err := contentHandler(ctx, c.path)
	if err != nil {
		return "", errors.Join(ErrGetContent, err)
	}
	return string(content), ctx.Err()
}

func (c *config) Source() string {
	return strings.Join([]string{string(c.protocol), c.path}, ProtoDelimiter)
}
