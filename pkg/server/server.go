package server

import (
	"net/http"

	"git.mylogic.dev/homelab/go-arcs/api/gen/proto/go/server/v1/serverv1connect"
	"git.mylogic.dev/homelab/go-arcs/pkg/mappings/collector"
	"git.mylogic.dev/homelab/go-arcs/pkg/mappings/config"
	"git.mylogic.dev/homelab/go-arcs/pkg/store"
	"github.com/grafana/alloy-remote-config/api/gen/proto/go/collector/v1/collectorv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Server struct {
	*http.Server
	configs    config.Store
	collectors collector.Store
}

func New(addr string, configs config.Store, collectors collector.Store) *Server {
	if configs == nil {
		configs = store.NewStore[config.Config](nil, nil)
	}

	if collectors == nil {
		collectors = store.NewStore[collector.Collector](nil, nil)
	}

	server := &Server{
		nil,
		configs,
		collectors,
	}

	mux := http.NewServeMux()
	mux.Handle(collectorv1connect.NewCollectorServiceHandler(server))
	mux.Handle(serverv1connect.NewCollectorManagerHandler(server))
	mux.Handle(serverv1connect.NewConfigManagerHandler(server))
	// Mount some handlers here.
	server.Server = &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
		// Don't forget timeouts!
	}

	return server
}
