package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andrei-cloud/gophermart/internal/config"
	repo "github.com/andrei-cloud/gophermart/internal/repo/inmem"
	"github.com/stretchr/testify/assert"
)

func Test_server_User(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		content string
		path    string
		body    io.Reader
		status  int
	}{
		{
			name:    "user login bad request",
			method:  "POST",
			content: "",
			path:    "/api/user/login",
			body:    nil,
			status:  http.StatusBadRequest,
		},
		{
			name:    "user login internal error",
			method:  "POST",
			content: "application/json",
			path:    "/api/user/login",
			body:    nil,
			status:  http.StatusInternalServerError,
		},
		{
			name:    "user login unauthorized",
			method:  "POST",
			content: "application/json",
			path:    "/api/user/login",
			body:    strings.NewReader(`{"login":"user","password":"1234"}`),
			status:  http.StatusUnauthorized,
		},
		{
			name:    "user register bad request",
			method:  "POST",
			content: "",
			path:    "/api/user/register",
			body:    nil,
			status:  http.StatusBadRequest,
		},
		{
			name:    "user register internal error",
			method:  "POST",
			content: "application/json",
			path:    "/api/user/register",
			body:    nil,
			status:  http.StatusInternalServerError,
		},
		{
			name:    "user register created",
			method:  "POST",
			content: "application/json",
			path:    "/api/user/register",
			body:    strings.NewReader(`{"login":"user","password":"1234"}`),
			status:  http.StatusOK,
		},
		{
			name:    "user register conflict",
			method:  "POST",
			content: "application/json",
			path:    "/api/user/register",
			body:    strings.NewReader(`{"login":"user","password":"1234"}`),
			status:  http.StatusConflict,
		},
		{
			name:    "user login OK",
			method:  "POST",
			content: "application/json",
			path:    "/api/user/login",
			body:    strings.NewReader(`{"login":"user","password":"1234"}`),
			status:  http.StatusOK,
		},
	}

	db := repo.NewInMemRepo()
	s := NewServer(&config.Cfg)
	s.WithDB(db).SetupRoutes()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			req := httptest.NewRequest(tt.method, tt.path, tt.body)
			req.Header.Set("Content-Type", tt.content)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()
			assert.Equal(tt.status, res.StatusCode)
		})
	}
}
