package config

import (
	"flag"
	"strings"

	"github.com/caarlos0/env"
	"github.com/rs/zerolog/log"
)

var Cfg Config

type Config struct {
	Address       string `env:"RUN_ADDRESS"`
	AccrualSystem string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DBURI         string `env:"DATABASE_URI"`
	Debug         bool   `env:"LOG_LEVEL" envDefault:"true"`
}

func init() {
	addressPtr := flag.String("a", "localhost:8080", "address to serve; format: host:port")
	AccrualPtr := flag.String("r", "http//localhost:9090", "address of accrual system format: host:port")
	dsnPtr := flag.String("d", "postgres://postgres:rootpassword@localhost:5432/gophermart", "database connection string")

	flag.Parse()
	Cfg = Config{}
	if err := env.Parse(&Cfg); err != nil {
		log.Fatal().AnErr("init", err)
	}
	if Cfg.Address == "" {
		Cfg.Address = *addressPtr
	}
	if Cfg.AccrualSystem == "" {
		Cfg.AccrualSystem = *AccrualPtr
	}
	if Cfg.DBURI == "" {
		Cfg.DBURI = *dsnPtr
	}

	el := strings.Split(Cfg.AccrualSystem, "//")
	if len(el) > 1 {
		Cfg.AccrualSystem = el[1]
	}
}
