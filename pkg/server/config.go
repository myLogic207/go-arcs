package server

import (
	"context"
	"errors"
	"log"
	"strings"

	"connectrpc.com/connect"
	serverv1 "git.mylogic.dev/homelab/go-arcs/api/gen/proto/go/server/v1"
	"git.mylogic.dev/homelab/go-arcs/pkg/store"
	collectorv1 "github.com/grafana/alloy-remote-config/api/gen/proto/go/collector/v1"
	"golang.org/x/sync/errgroup"
)

func (s *Server) GetConfig(
	ctx context.Context,
	req *connect.Request[collectorv1.GetConfigRequest],
) (*connect.Response[collectorv1.GetConfigResponse], error) {
	id := req.Msg.GetId()
	currentHash := req.Msg.GetHash()
	attributes := req.Msg.GetLocalAttributes()

	log.Printf("Client %v send configuration request %+v", id, attributes)
	configs := s.configs.GetByAttributes(ctx, attributes)

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
	resolvedConfig := strings.Join(results, "\n")
	newHash := store.Hash([]byte(resolvedConfig))
	// globalStorage.Set(collector.id, resolvedConfig.String())
	return connect.NewResponse(&collectorv1.GetConfigResponse{
		Content:     resolvedConfig,
		Hash:        newHash,
		NotModified: currentHash == newHash,
	}), nil
}

func (s *Server) ListConfigs(
	ctx context.Context,
	req *connect.Request[serverv1.ListRequest],
	stream *connect.ServerStream[serverv1.GetConfigResponse],
) error {
	attributes := req.Msg.GetLocalAttributes()
	log.Printf("List config request received %+v", attributes)
	configs := s.configs.GetByAttributes(ctx, attributes)
	for _, config := range configs {
		stream.Send(&serverv1.GetConfigResponse{
			Source:          config.Source(),
			LocalAttributes: config.Attributes(),
		})
	}
	return nil
}
