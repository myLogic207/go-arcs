package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

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
		"log": {
			Name:  "log",
			Value: "console",
			Message: `Path to log to, if path is dir 'server.log' is appended.
If Exists, will rotate. If an error occurs, console is used as fallback.
Special value 'console' will force console.`,
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
			log.Println("System stopped")
			done <- true
		case sig := <-sigs:
			log.Printf("Received signal %v", sig)
			done <- true
		}
	}
}

func getLogFile(name string) (*os.File, error) {
	if name == "console" {
		return os.Stdout, errors.New("forced console logging")
	}

	logPath, err := filepath.Abs(name)
	if err != nil {
		return os.Stdout, err
	}

	stat, err := os.Stat(logPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return os.Stdout, err
	}
	err = nil
	if stat.IsDir() {
		logPath = filepath.Join(logPath, "server.log")
		stat, err = os.Stat(logPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return os.Stdout, err
		}
	}

	if err == nil && stat.Size() > 0 {
		logPathPart, ok := strings.CutSuffix(logPath, ".log")
		//Reference Time is: 01/02 03:04:05PM 2006 MST = January 2, 2006 at 3:04:05 PM MST
		timestamp := time.Now().Format("2006-01-02.03-04-05")
		var logPathNew string
		if ok {
			logPathNew = fmt.Sprintf("%v.%v.log", logPathPart, timestamp)
		} else {
			logPathNew = fmt.Sprintf("%v.%v", logPathPart, timestamp)
		}
		// rotate old log
		if err := os.Rename(logPath, logPathNew); err != nil {
			return nil, err
		}
	}
	return os.Create(logPath)
}

func main() {
	log.Print("Starting...")
	mainCtx := context.Background()
	ctx, cancel := context.WithCancel(mainCtx)
	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	go cleanup(ctx, cancel, sigs, done)

	flags, _ := args.Init(customFlags)
	logFile, err := getLogFile(*flags["log"].(*string))
	if err != nil {
		log.Println("Log (file) path not found/readable, falling back to console:", err)
		logFile = os.Stdout
	} else {
		defer logFile.Close()
	}
	log.SetOutput(logFile)

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

	log.Print("Starting Server")
	go func() {
		err := s.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen failed: %v", err)
		}
	}()
	log.Printf("Server listening on %v and ready to accept connections", address)

	<-done

	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Failed to stop Server: %v", err)
	}
	// listener.Close()
	cancel()
}
