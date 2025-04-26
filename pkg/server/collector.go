package server

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	serverv1 "git.mylogic.dev/homelab/go-arcs/api/gen/proto/go/server/v1"
	"git.mylogic.dev/homelab/go-arcs/pkg/mappings/collector"
	collectorv1 "github.com/grafana/alloy-remote-config/api/gen/proto/go/collector/v1"
)

var (
	ErrCollectorAdd    = errors.New("could not add collector")
	ErrCollectorExists = errors.New("ID is already registered")
	ErrCollectorRemove = errors.New("could not remove collector")
)

func (s *Server) GetCollector(
	ctx context.Context,
	req *connect.Request[serverv1.GetCollectorRequest],
) (*connect.Response[serverv1.GetCollectorsResponse], error) {
	logRequest(req)
	id := req.Msg.GetId()
	col := s.collectors.Get(ctx, id)
	return connect.NewResponse(&serverv1.GetCollectorsResponse{
		Id:              col.ID(),
		LocalAttributes: col.Attributes(),
		Name:            col.Name(),
	}), nil
}

func (s *Server) ListCollectors(
	ctx context.Context,
	req *connect.Request[serverv1.ListRequest],
	stream *connect.ServerStream[serverv1.GetCollectorsResponse],
) error {
	logRequest(req)
	attributes := req.Msg.GetLocalAttributes()

	var collectors []collector.Collector
	if attributes != nil || len(attributes) > 0 {
		collectors = s.collectors.GetByAttributes(ctx, attributes)
	} else {
		collectors = s.collectors.List(ctx)
	}

	for _, col := range collectors {
		stream.Send(&serverv1.GetCollectorsResponse{
			Id:              col.ID(),
			LocalAttributes: col.Attributes(),
			Name:            col.Name(),
		})
	}
	return nil
}

func (s *Server) RegisterCollector(
	ctx context.Context,
	req *connect.Request[collectorv1.RegisterCollectorRequest],
) (*connect.Response[collectorv1.RegisterCollectorResponse], error) {
	logRequest(req)
	col := collector.New(
		req.Msg.GetId(),
		req.Msg.GetName(),
		req.Msg.GetLocalAttributes(),
		"",
	)

	if s.collectors.Get(ctx, col.ID()) != nil {
		return nil, ErrCollectorExists
	}
	_, err := s.collectors.Set(ctx, col)
	if err != nil {
		return nil, errors.Join(ErrCollectorAdd, err)
	}
	return connect.NewResponse(&collectorv1.RegisterCollectorResponse{}), nil
}

func (s *Server) UnregisterCollector(
	ctx context.Context,
	req *connect.Request[collectorv1.UnregisterCollectorRequest],
) (*connect.Response[collectorv1.UnregisterCollectorResponse], error) {
	logRequest(req)
	collectorID := req.Msg.GetId()

	_, err := s.collectors.Remove(ctx, collectorID)
	if err != nil {
		return nil, errors.Join(ErrCollectorRemove, err)
	}

	return connect.NewResponse(&collectorv1.UnregisterCollectorResponse{}), nil
}
