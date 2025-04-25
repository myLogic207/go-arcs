package config

import (
	"context"
	"errors"
	"slices"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

var (
	ErrLoadConfig = errors.New("cloud not load config")
	ErrConfFields = errors.New("config entry does not contain needed fields 'source' and 'attribute'")
	ErrConfCreate = errors.New("could not create config")
)

type Parser[t any] func([]byte) ([]t, error)

type configRaw struct {
	Source     string            `yaml:"source"`
	Attributes map[string]string `yaml:"attributes"`
}

// Load the Server configuration mappings from a file or directory
func Load(ctx context.Context, path string) ([]Config, error) {
	files, err := GetFileOrFiles(path)
	if err != nil {
		return nil, errors.Join(ErrLoadConfig, err)
	}
	eg, eCtx := errgroup.WithContext(ctx)
	configs := make([][]Config, len(files))

	for i, file := range files {
		eg.Go(func() error {
			content, err := fileHandler(eCtx, file)
			if err != nil {
				return nil
			}
			configs[i], err = ParseConfig(content)
			return err
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, errors.Join(ErrLoadConfig, err)
	}
	return slices.Concat(configs...), nil
}

func ParseConfig(content []byte) ([]Config, error) {
	var rawConfigs []configRaw
	err := yaml.Unmarshal(content, &rawConfigs)
	if err != nil {
		return nil, err
	}

	configs := make([]Config, len(rawConfigs))
	var errs error
	for i, conf := range rawConfigs {
		conf, err := New(conf.Source, conf.Attributes)
		if err != nil {
			errs = errors.Join(errs, err)
		}
		configs[i] = conf
	}

	return configs, errs
}
