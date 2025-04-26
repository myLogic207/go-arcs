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

func (s *Server) GetConfig(
	ctx context.Context,
	req *connect.Request[collectorv1.GetConfigRequest],
) (*connect.Response[collectorv1.GetConfigResponse], error) {
	logRequest(req)
	// id := req.Msg.GetId()
	attributes := req.Msg.GetLocalAttributes()
	configs := s.configs.GetByAttributes(ctx, attributes)
	collectorID := req.Msg.GetId()

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
	// check if collector is registered

	eg, getCtx := errgroup.WithContext(ctx)
	results := make([]string, len(configs))
	for i, config := range configs {
		if err := ctx.Err(); err != nil {
			getErr := eg.Wait()
			return nil, errors.Join(err, getErr)
		}
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
		return nil, err
	}
	resolvedConfig := strings.Join(results, " ")
	newHash := store.Hash([]byte(resolvedConfig))
	modified := currentHash == newHash
	if modified {
		collector.SetHash(newHash)
	}
	// globalStorage.Set(collector.id, resolvedConfig.String())
	return connect.NewResponse(&collectorv1.GetConfigResponse{
		Content:     resolvedConfig,
		Hash:        newHash,
		NotModified: modified,
	}), nil
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
