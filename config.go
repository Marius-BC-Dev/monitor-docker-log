package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	FlagCfg = "cfg"
)

type Config struct {
	DockerVersion string          `json:"docker_version"`
	Containers    []ContainerInfo `json:"containers"`
	Tail          string          `json:"tail"`
	HookUrl       string          `json:"hook_url"`
}

type ContainerInfo struct {
	ContainerID string `json:"container_id"`
	ServerName  string `json:"server_name"`
	HookUrl     string `json:"hook_url"`
}

type LogConfig struct {
	// Environment defining the log format ("production" or "development").
	Environment LogEnvironment `mapstructure:"Environment"`
	// Level of log, e.g. INFO, WARN, ...
	Level string `mapstructure:"Level"`
	// Outputs
	Outputs []string `mapstructure:"Outputs"`
}

func load(ctx *cli.Context) (*Config, error) {
	cfgPath := ctx.String(FlagCfg)
	if cfgPath != "" {
		f, err := os.Open(cfgPath) //nolint:gosec
		if err != nil {
			return nil, err
		}
		defer func() {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}()

		b, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		var cfg Config
		err = json.Unmarshal([]byte(b), &cfg)
		if err != nil {
			return nil, err
		}

		return &cfg, nil

	}

	return nil, errors.New("need config")
}
