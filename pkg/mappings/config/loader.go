package config

import (
	"context"
	"errors"
	"os"
	"path"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

var (
	ErrLoadConfig = errors.New("cloud not load config")
	ErrConfFields = errors.New("config entry does not contain needed fields 'source' and 'attribute'")
	ErrConfCreate = errors.New("could not create config")
)

type configRaw struct {
	Source     string            `yaml:"source"`
	Attributes map[string]string `yaml:"attributes"`
}

// Load the Server configuration mappings from a file or directory
func Load(ctx context.Context, path string) ([]Config, error) {
	// found out if path is folder or file
	stat, err := os.Stat(path)
	if err != nil {
		return nil, errors.Join(ErrLoadConfig, err)
	}

	files := make(chan string, 1)
	if stat.IsDir() {
		// load files in parallel
		go loadDir(ctx, path, files)
	} else {
		// load single file directly and store in channel
		files <- path
		close(files)
	}
	return parseConfigs(ctx, files)

}

func parseConfigs(ctx context.Context, files <-chan string) ([]Config, error) {
	collector := make(chan Config)
	eg, eCtx := errgroup.WithContext(ctx)

	for file := range files {
		if err := eCtx.Err(); err != nil {
			return nil, err
		}

		eg.Go(func() error {
			configs, err := loadAndParseFile(eCtx, file)
			if err != nil {
				return err
			}
			for _, config := range configs {
				select {
				case <-eCtx.Done():
					return eCtx.Err()
				case collector <- config:
					continue
				}
			}
			return nil
		})
	}

	entries := []Config{}
	eg.Go(func() error {
		var err error
		entries, err = collectConfigs(eCtx, collector)
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return entries, nil
}

func collectConfigs(ctx context.Context, stream <-chan Config) ([]Config, error) {
	var configs []Config

	for conf := range stream {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		configs = append(configs, conf)
	}

	return configs, nil
}

// expects a valid dirpath!
func loadDir(ctx context.Context, dir string, files chan<- string) error {
	defer close(files)
	dirContent, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, ent := range dirContent {
		if ent.IsDir() {
			continue
		}
		file := path.Join(dir, ent.Name())
		select {
		case <-ctx.Done():
			return ctx.Err()
		case files <- file:
		}
	}
	return nil
}

func loadAndParseFile(ctx context.Context, filepath string) ([]Config, error) {
	raw, err := fileHandler(ctx, filepath)
	if err != nil {
		return nil, err
	}
	content, err := parseRawConfig(raw)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func parseRawConfig(content []byte) ([]Config, error) {
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

	return configs, err
}
