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
}

type ContainerInfo struct {
	ContainerID string `json:"container_id"`
	ServerName  string `json:"server_name"`
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
