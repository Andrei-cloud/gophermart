package server

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth/v5"
)

var TokenAuth *jwtauth.JWTAuth

const password string = "my_secret"

func init() {
	TokenAuth = jwtauth.New("HS256", []byte(password), nil)
}

func (s *server) SetupRoutes() {
	s.router = chi.NewRouter()

	s.router.Use(Compressor)
	//Public routes
	s.router.Group(func(r chi.Router) {
		r.Post("/api/user/register", s.userRegister())
		r.Post("/api/user/login", s.userLogin())
		//authentication required handlers
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(TokenAuth))
			r.Use(jwtauth.Authenticator)
			r.Post("/api/user/orders", s.userAddOrder())
			r.Get("/api/user/orders", s.userOrderList())
			r.Get("/api/user/balance", s.userBalance())
			r.Post("/api/user/balance/withdraw", s.userWithdraw())
			r.Get("/api/user/withdrawals", s.userWithdrawalList())
		})
	})

	//private routes

	s.Server.Handler = s.router
}
