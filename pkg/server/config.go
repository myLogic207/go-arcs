package server

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	serverv1 "git.mylogic.dev/homelab/go-arcs/api/gen/proto/go/server/v1"
	"git.mylogic.dev/homelab/go-arcs/pkg/mappings/config"
	"git.mylogic.dev/homelab/go-arcs/pkg/store"
	collectorv1 "github.com/grafana/alloy-remote-config/api/gen/proto/go/collector/v1"
	"golang.org/x/sync/errgroup"
)

var (
	ErrGetConfig = errors.New("failed to parse config")
)

func (s *Server) GetConfig(
	ctx context.Context,
	req *connect.Request[collectorv1.GetConfigRequest],
) (*connect.Response[collectorv1.GetConfigResponse], error) {
	logRequest(req)
	// id := req.Msg.GetId()
	attributes := req.Msg.GetLocalAttributes()
	collectorID := req.Msg.GetId()

	// check if collector is registered
	collector := s.collectors.Get(ctx, collectorID)
	if collector == nil {
		return nil, ErrCollectorNotRegistered
	}

	currentHash := ""
	if reqHash := req.Msg.GetHash(); reqHash != "" {
		currentHash = reqHash
	} else if hash := collector.GetHash(); hash != "" {
		currentHash = hash
	} else if hash == "" && reqHash != "" {
		collector.SetHash(reqHash)
	}

	configs := s.configs.GetByAttributes(ctx, attributes)
	config, err := getCollectorConfig(ctx, configs)
	if err != nil {
		return nil, errors.Join(ErrGetConfig)
	}
	newHash := store.Hash([]byte(config))
	modified := currentHash == newHash
	if modified {
		collector.SetHash(newHash)
	}

	return connect.NewResponse(&collectorv1.GetConfigResponse{
		Content:     config,
		Hash:        newHash,
		NotModified: modified,
	}), nil
}

func getCollectorConfig(
	ctx context.Context,
	configs []config.Config,
) (string, error) {
	eg, getCtx := errgroup.WithContext(ctx)
	results := make([]string, len(configs))
	for i, config := range configs {
		i, config := i, config // https://golang.org/doc/faq#closures_and_goroutines
		eg.Go(func() error {
			content, err := config.Content(getCtx)
			if err == nil {
				results[i] = content
			}
			return err
		})
	}
	if err := eg.Wait(); err != nil {
		return "", err
	}
	return strings.Join(results, " "), nil
}

func (s *Server) ListConfigs(
	ctx context.Context,
	req *connect.Request[serverv1.ListRequest],
	stream *connect.ServerStream[serverv1.GetConfigResponse],
) error {
	logRequest(req)
	attributes := req.Msg.GetLocalAttributes()

	var configs []config.Config
	if attributes != nil || len(attributes) > 0 {
		configs = s.configs.GetByAttributes(ctx, attributes)
	} else {
		// if no attributes specified, return all configs
		configs = s.configs.List(ctx)
	}

	for _, config := range configs {
		stream.Send(&serverv1.GetConfigResponse{
			Source:          config.Source(),
			LocalAttributes: config.Attributes(),
		})
	}
	return nil
}
