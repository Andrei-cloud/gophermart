package server

import (
	"net/http"
	"time"

	"github.com/andrei-cloud/gophermart/internal/config"
	"github.com/andrei-cloud/gophermart/internal/repo"
	"github.com/go-chi/chi"
)

type server struct {
	http.Server

	db     repo.Repository
	router *chi.Mux
}

func NewServer(cfg *config.Config) *server {
	return &server{
		Server: http.Server{
			Addr:           ":8080",
			ReadTimeout:    60 * time.Second,
			WriteTimeout:   60 * time.Second,
			IdleTimeout:    30 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		db: nil,
	}
}

func (s *server) WithDB(r repo.Repository) *server {
	s.db = r
	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
