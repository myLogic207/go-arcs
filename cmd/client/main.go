package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	serverv1 "git.mylogic.dev/homelab/go-arcs/api/gen/proto/go/server/v1"
	"git.mylogic.dev/homelab/go-arcs/api/gen/proto/go/server/v1/serverv1connect"
	"git.mylogic.dev/homelab/go-arcs/internal/args"
	collectorv1 "github.com/grafana/alloy-remote-config/api/gen/proto/go/collector/v1"
	"github.com/grafana/alloy-remote-config/api/gen/proto/go/collector/v1/collectorv1connect"
)

var (
	customFlags = map[string]args.Flag{
		"port": {
			Name:    "port",
			Value:   8080,
			Message: "The Port the Server binds to",
		},
		"addr": {
			Name:    "host",
			Value:   "172.17. 0.1",
			Message: "The IP Address to bind to, if none specified uses default docker host address",
		},
		// args.Flag{
		// 	Name:    "validate",
		// 	Value:   false,
		// 	Message: "Flag to validate the configuration and the stop",
		// },
	}
)

type action byte

const (
	listConfigs = action(0x11)
	getConfig   = action(0x12)
	// addConfig    = action(0x13)
	// removeConfig    = action(0x14)
	listCollectors  = action(0x21)
	getCollector    = action(0x22)
	addCollector    = action(0x23)
	removeCollector = action(0x24)
)

func parseArgs(names []string) (action, []string) {
	var raw uint8
	if len(names) < 2 {
		log.Fatal("Command incorrect. Usage: [scope] [action] [attributes]")
	}
	switch names[0] {
	case "config":
		raw += 0x10
	case "configs":
		raw += 0x10
	case "collector":
		raw += 0x20
	case "collectors":
		raw += 0x20
	}

	switch names[1] {
	case "list":
		raw += 0x01
	case "get":
		raw += 0x02
	case "add":
		raw += 0x03
	case "remove":
		raw += 0x04
	}
	if raw <= 0x10 {
		log.Fatalf("No known action '%v' for '%v", names[1], names[0])
		return 0, nil
	}

	return action(raw), names[2:]
}

// command line attributes take the form of key=value,key2=value2
func parseAttributes(raw string) (map[string]string, error) {
	pairs := strings.Split(raw, ",")
	attributes := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("could not parse %v as attribute", pair)
		}
		attributes[key] = value
	}
	return attributes, nil
}

func main() {
	ctx := context.Background()
	flags, unnamed := args.Init(customFlags)
	address := fmt.Sprintf("http://%v:%v", *flags["addr"].(*string), *flags["port"].(*int))
	configClient := serverv1connect.NewConfigManagerClient(
		http.DefaultClient,
		address,
	)
	collectorClient := collectorv1connect.NewCollectorServiceClient(
		http.DefaultClient,
		address,
	)
	collectorClientAddon := serverv1connect.NewCollectorManagerClient(
		http.DefaultClient,
		address,
	)

	action, rawArguments := parseArgs(unnamed)
	log.Printf("calling %v, executing %v with %+v", address, action, rawArguments)
	switch action {
	case listConfigs:
		attributes := make(map[string]string)
		if len(rawArguments) >= 1 {
			var err error
			attributes, err = parseAttributes(rawArguments[0])
			if err != nil {
				log.Fatal(err)
			}
		}
		res, err := configClient.ListConfigs(
			ctx,
			connect.NewRequest(&serverv1.ListRequest{
				LocalAttributes: attributes,
			}),
		)
		if err != nil {
			log.Fatal(err)
		}
		// log.Print(res.Msg().GetSource())
		for res.Receive() {
			log.Print(res.Msg().GetSource())
		}
	case getConfig:
		attributes, err := parseAttributes(rawArguments[0])
		if err != nil {
			log.Fatal(err)
		}
		res, err := collectorClient.GetConfig(
			ctx,
			connect.NewRequest(&collectorv1.GetConfigRequest{
				Id:              "ARCS-Client",
				LocalAttributes: attributes,
			}),
		)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%v: modified %v\n%v", res.Msg.GetHash(), res.Msg.GetNotModified(), res.Msg.GetContent())
		// case addConfig:
		// case removeConfig:
	case listCollectors:
		attributes := make(map[string]string)
		if len(rawArguments) >= 1 {
			var err error
			attributes, err = parseAttributes(rawArguments[0])
			if err != nil {
				log.Fatal(err)
			}
		}
		res, err := collectorClientAddon.ListCollectors(
			ctx,
			connect.NewRequest(&serverv1.ListRequest{
				LocalAttributes: attributes,
			}),
		)
		if err != nil {
			log.Fatal(err)
		}
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%v: %v\n(%+v)", res.Msg().GetId(), res.Msg().GetName(), res.Msg().GetLocalAttributes())
		for res.Receive() {
			log.Printf("%v: %v\n(%+v)", res.Msg().GetId(), res.Msg().GetName(), res.Msg().GetLocalAttributes())
		}
	case getCollector:
	case addCollector:
	case removeCollector:
	}
}
