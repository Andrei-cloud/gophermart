package config

import (
	"flag"
	"strings"
	"sync"

	"github.com/caarlos0/env"
	"github.com/rs/zerolog/log"
)

var (
	cfg  Config
	once sync.Once
)

type Config struct {
	Address       string `env:"RUN_ADDRESS"`
	AccrualSystem string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DBURI         string `env:"DATABASE_URI"`
	Debug         bool   `env:"LOG_LEVEL" envDefault:"true"`
}

func GetConfig() *Config {
	once.Do(func() {
		addressPtr := flag.String("a", "localhost:8080", "address to serve; format: host:port")
		AccrualPtr := flag.String("r", "http//localhost:9090", "address of accrual system format: host:port")
		dsnPtr := flag.String("d", "postgres://postgres:rootpassword@localhost:5432/gophermart", "database connection string")

		flag.Parse()
		cfg = Config{}
		if err := env.Parse(&cfg); err != nil {
			log.Fatal().AnErr("init", err)
		}
		if cfg.Address == "" {
			cfg.Address = *addressPtr
		}
		if cfg.AccrualSystem == "" {
			cfg.AccrualSystem = *AccrualPtr
		}
		if cfg.DBURI == "" {
			cfg.DBURI = *dsnPtr
		}

		el := strings.Split(cfg.AccrualSystem, "//")
		if len(el) > 1 {
			cfg.AccrualSystem = el[1]
		}
	})

	return &cfg
}
