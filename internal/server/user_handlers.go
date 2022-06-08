package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/andrei-cloud/gophermart/internal/domain"
	repo "github.com/andrei-cloud/gophermart/internal/repo"
	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog/log"
)

func isValidType(w http.ResponseWriter, r *http.Request, expectedType string) bool {
	ct := r.Header.Get("Content-Type")
	if ct != expectedType {
		log.Debug().Msg("isValidType: invalid content type")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return false
	}
	return true
}

func generateToken(w http.ResponseWriter, r *http.Request, userID int64) error {
	claims := map[string]interface{}{"userId": userID}

	jwtauth.SetIssuedAt(claims, time.Now())
	jwtauth.SetExpiryIn(claims, 24*time.Hour)

	token, tokenString, err := TokenAuth.Encode(claims)
	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Path:     "/",
		Domain:   "localhost",
		Expires:  token.Expiration(),
		Secure:   false,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	return nil
}

func (s *server) userLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isValidType(w, r, "application/json") {
			return
		}

		user := domain.UserModel{}
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			log.Debug().AnErr("userLogin: decode request", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		userID, err := user.Login(s.db)
		if err != nil {
			log.Debug().AnErr("userLogin: login", err)
			switch errors.Is(err, repo.ErrNotExists) {
			case true:
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		err = generateToken(w, r, userID)
		if err != nil {
			log.Debug().AnErr("login: encode token", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s *server) userRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isValidType(w, r, "application/json") {
			return
		}

		user := domain.UserModel{}
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			log.Debug().AnErr("userRegister: decode body", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		userID, err := user.Register(s.db)
		if err != nil {
			log.Debug().AnErr("userRegister: register", err)
			switch errors.Is(err, repo.ErrAlreadyExists) {
			case true:
				http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		err = generateToken(w, r, userID)
		if err != nil {
			log.Debug().AnErr("register: encode token", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s *server) userBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			log.Debug().AnErr("userBalance: jwt form context", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		userID, ok := claims["userId"].(float64)
		if !ok {
			log.Debug().AnErr("userBalance: not valid claims", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		user := domain.UserModel{ID: int64(userID)}

		balance, err := user.GetBalance(s.db)
		if err != nil {
			log.Debug().AnErr("userBalance: get balance", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(&balance)
		if err != nil {
			log.Debug().AnErr("userBalance: encoding response", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
