package main

import (
	"errors"
	"flag"
	"os"

	configpkg "github.com/bazueva/gofermart/cmd/config"
	"github.com/caarlos0/env/v11"

	"go.uber.org/zap"
)

type config struct {
	ServerAddr  configpkg.ServerAddr `env:"RUN_ADDRESS"`
	DatabaseDSN string               `env:"DATABASE_URI"`
	SecretKey   string               `env:"SECRET_KEY"`

	logger *zap.Logger
}

func readConfig() (config, error) {
	cfg := config{
		ServerAddr: configpkg.ServerAddr{
			Host: "localhost",
			Port: 8080,
		},
	}

	err := parseFlags(&cfg)
	if err != nil {
		return config{}, err
	}

	err = env.Parse(&cfg)
	if err != nil {
		return cfg, err
	}

	if cfg.SecretKey == "" {
		return config{}, errors.New("secret_key required")
	}

	return cfg, nil
}

func parseFlags(config *config) error {
	serverFlags := flag.NewFlagSet("", flag.ContinueOnError)
	serverFlags.Var(&config.ServerAddr, "a", "address http server")
	serverFlags.StringVar(&config.DatabaseDSN, "d", "", "Database DSN")
	serverFlags.StringVar(&config.SecretKey, "s", "", "Secret Key")

	if len(os.Args) > 1 {
		err := serverFlags.Parse(os.Args[1:])
		if err != nil {
			return err
		}
	}

	return nil
}
