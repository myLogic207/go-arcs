package args

import (
	"flag"
	"log"
)

type Flag struct {
	Name    string
	Value   any
	Message string
}

func Init(
	flags map[string]Flag,
) (map[string]any, []string) {
	args := make(map[string]any)
	for name, rawFlag := range flags {
		switch value := rawFlag.Value.(type) {
		case int:
			arg := flag.Int(rawFlag.Name, value, rawFlag.Message)
			args[name] = arg
		case string:
			arg := flag.String(rawFlag.Name, value, rawFlag.Message)
			args[name] = arg
		case bool:
			arg := flag.Bool(rawFlag.Name, value, rawFlag.Message)
			args[name] = arg
		default:
			log.Printf("Unknown type for value %v detected, no arg registered", value)
		}
	}
	if flag.Parsed() {
		return nil, nil
	}
	flag.Parse()

	return args, flag.Args()
}
