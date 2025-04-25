package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"git.mylogic.dev/homelab/go-arcs/internal/args"
	"git.mylogic.dev/homelab/go-arcs/pkg/mappings/collector"
	"git.mylogic.dev/homelab/go-arcs/pkg/mappings/config"
	"git.mylogic.dev/homelab/go-arcs/pkg/server"
	"git.mylogic.dev/homelab/go-arcs/pkg/store"
)

var (
	customFlags = map[string]args.Flag{
		"config": {
			Name:    "config",
			Value:   "mappings.yaml",
			Message: "Specify the path to a config file or folder (includes all yml|yaml files)",
		},
		"port": {
			Name:    "port",
			Value:   8080,
			Message: "The Port the Server binds to",
		},
		"addr": {
			Name:    "ip",
			Value:   "0.0.0.0",
			Message: "The IP Address to bind to, if none specified, binds to all ipv4",
		},
		// args.Flag{
		// 	Name:    "validate",
		// 	Value:   false,
		// 	Message: "Flag to validate the configuration and the stop",
		// },
	}
)

func cleanup(ctx context.Context, cancel context.CancelFunc, sigs chan os.Signal, done chan<- bool) {
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ctx.Done():
			cancel()
			fmt.Println("System stopped")
			done <- true
		case sig := <-sigs:
			fmt.Println(sig)
			done <- true
		}
	}
}

func main() {
	log.Print("Starting...")
	mainCtx := context.Background()
	ctx, cancel := context.WithCancel(mainCtx)
	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	go cleanup(ctx, cancel, sigs, done)

	flags, _ := args.Init(customFlags)
	configPath := flags["config"].(*string)
	log.Printf("Loading configs from %v", *configPath)
	initConfigs, err := config.Load(ctx, *configPath)
	if err != nil {
		cancel()
		log.Fatal(err)
	}
	log.Printf("Loaded %v configs, creating store", len(initConfigs))
	initConfigStore := store.NewStore[config.Config](nil, nil)
	if _, err := initConfigStore.Load(ctx, initConfigs); err != nil {
		cancel()
		log.Fatal(err)
	}

	log.Print("Created init config store")

	collectorStore := store.NewStore(
		store.ObjectStore[collector.Collector]{},
		store.MappingStore{},
	)
	log.Print("Created collector store")

	address := fmt.Sprintf("%v:%v", *flags["addr"].(*string), *flags["port"].(*int))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		cancel()
		log.Fatal(err)
	}
	s := server.New(address, initConfigStore, collectorStore)

	log.Print("Starting Server...")
	go func() {
		err := s.Serve(listener)
		if err != nil {
			log.Fatalf("listen failed: %v", err)
		}
	}()

	<-done

	s.Shutdown(ctx)
	cancel()
	log.Print("Stopped")
}
