package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andrei-cloud/gophermart/internal/config"
	"github.com/andrei-cloud/gophermart/internal/repo"
	"github.com/andrei-cloud/gophermart/internal/repo/indb"
	"github.com/andrei-cloud/gophermart/internal/repo/inmem"
	"github.com/andrei-cloud/gophermart/internal/server"
	"github.com/andrei-cloud/gophermart/internal/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.GetConfig()
	var (
		db  repo.Repository
		err error
	)

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("DEBUG LEVEL IS ENABLED")

	}

	log.Info().Msg("Starting...")

	if cfg.DBURI == "" {
		db = inmem.NewInMemRepo()
	} else {
		db = indb.NewDB(cfg.DBURI)
	}

	s := server.NewServer(cfg)

	s.WithDB(db).SetupRoutes()

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal().Msg("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := s.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		serverStopCtx()
	}()

	// launch worker
	wrkr := worker.NewWorker(cfg.AccrualSystem, db)

	go wrkr.Run(serverCtx)

	// Run the server
	err = s.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal().Msg(err.Error())
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	log.Info().Msg("Stopped...")
}
