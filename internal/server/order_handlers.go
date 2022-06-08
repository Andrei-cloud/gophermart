package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/andrei-cloud/gophermart/internal/domain"
	repo "github.com/andrei-cloud/gophermart/internal/repo"
	"github.com/andrei-cloud/gophermart/pkg/utils"
	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog/log"
)

func (s *server) userAddOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isValidType(w, r, "text/plain") {
			return
		}

		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		userID, ok := claims["userId"].(float64)
		if !ok {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		order := domain.OrderModel{}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Debug().AnErr("userAddOrder: reading body", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		order.Number = string(b)
		if !utils.IsValidLuhn(order.Number) {
			log.Debug().AnErr("userAddOrder: isvalid luhn number", fmt.Errorf("invalid luhn"))
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}

		order.UserID = int64(userID)

		err = order.Register(s.db)
		if err != nil {
			log.Debug().AnErr("userAddOrder: register", err)
			if errors.Is(err, domain.ErrDontMatch) {
				http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
				return
			} else if errors.Is(err, repo.ErrAlreadyExists) {
				http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
				return
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func (s *server) userOrderList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		userID, ok := claims["userId"].(float64)
		if !ok {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		order := domain.OrderModel{
			UserID: int64(userID),
		}
		list, err := order.CreditList(s.db)
		if err != nil {
			log.Debug().AnErr("userOrderList: credit list", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if len(list) == 0 {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(&list)
			if err != nil {
				log.Debug().AnErr("userOrderList: encoding response", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
	}
}

func (s *server) userWithdraw() http.HandlerFunc {
	request := struct {
		userID int64
		Order  string  `json:"order"`
		Value  float64 `json:"sum"`
	}{}
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		if !isValidType(w, r, "application/json") {
			return
		}

		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		userID, ok := claims["userId"].(float64)
		if !ok {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		request.userID = int64(userID)

		err = json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Debug().AnErr("userWithdraw: decoding request body", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		if !utils.IsValidLuhn(request.Order) {
			log.Debug().AnErr("userWithdraw: isvalid luhn number", fmt.Errorf("invalid luhn"))
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}

		order := domain.OrderModel{
			UserID: request.userID,
			Number: request.Order,
			Value:  request.Value,
		}

		err = order.Withdraw(s.db)
		if err != nil {
			log.Debug().AnErr("userWithdraw: withdraw", err)
			if errors.Is(err, domain.ErrIsufficientFunds) {
				http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
				return
			} else if errors.Is(err, repo.ErrAlreadyExists) {
				http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
				return
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s *server) userWithdrawalList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		userID, ok := claims["userId"].(float64)
		if !ok {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		order := domain.OrderModel{
			UserID: int64(userID),
		}

		list, err := order.DebitList(s.db)
		if err != nil {
			log.Debug().AnErr("userWithdrawalList: debit list", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if len(list) == 0 {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(&list)
			if err != nil {
				log.Debug().AnErr("userWithdrawalList: encoding response", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
	}
}
