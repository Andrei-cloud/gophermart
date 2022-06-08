package domain

import (
	"errors"
	"time"

	"github.com/andrei-cloud/gophermart/internal/repo"
)

var (
	ErrDontMatch        = errors.New("user dont match")
	ErrIsufficientFunds = errors.New("issuficient funds")
)

type OrderModel struct {
	UserID     int64   `json:"-"`
	Number     string  `json:"number,omitempty"`
	Status     string  `json:"status,omitempty"`
	Value      float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at,omitempty"`
}

func (o *OrderModel) Register(r repo.Repository) error {
	order, err := r.OrderGet(o.Number)
	if err != nil && !errors.Is(err, repo.ErrNotExists) {
		return err
	} else if errors.Is(err, repo.ErrNotExists) {
		order := repo.Order{
			Order:      o.Number,
			Type:       repo.CREDIT,
			UserID:     o.UserID,
			Status:     repo.NEW,
			UploadedAt: time.Now(),
		}
		_, err = r.OrderCreate(&order)
		if err != nil {
			return err
		}
		return nil
	}
	if order.UserID != o.UserID {
		return ErrDontMatch
	}

	return repo.ErrAlreadyExists
}

func (o *OrderModel) Withdraw(r repo.Repository) error {
	_, err := r.OrderGet(o.Number)
	if err != nil && !errors.Is(err, repo.ErrNotExists) {
		return err
	} else if errors.Is(err, repo.ErrNotExists) {
		user, err := r.UserGetByID(o.UserID)
		if err != nil {
			return err
		}

		if user.Balance < o.Value {
			return ErrIsufficientFunds
		}

		order := repo.Order{
			Order:      o.Number,
			Type:       repo.DEBIT,
			Value:      o.Value,
			UserID:     o.UserID,
			UploadedAt: time.Now(),
		}
		_, err = r.OrderCreate(&order)
		if err != nil {
			return err
		}

		user.Balance -= o.Value
		user.Withdrawal += o.Value

		err = r.UserUpdate(user)
		if err != nil {
			return err
		}

		return nil
	}

	return repo.ErrAlreadyExists
}

func (o *OrderModel) CreditList(r repo.Repository) ([]OrderModel, error) {
	orders, err := r.OrderGetList(o.UserID, repo.CREDIT)
	if err != nil {
		return nil, err
	}
	list := make([]OrderModel, 0)
	for _, order := range orders {
		listitem := &OrderModel{
			UserID:     order.UserID,
			Number:     order.Order,
			Status:     string(order.Status),
			Value:      order.Value,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		}
		list = append(list, *listitem)
	}
	return list, nil
}

func (o *OrderModel) DebitList(r repo.Repository) ([]OrderModel, error) {
	orders, err := r.OrderGetList(o.UserID, repo.DEBIT)
	if err != nil {
		return nil, err
	}
	list := make([]OrderModel, 0)
	for _, order := range orders {
		listitem := &OrderModel{
			UserID:     order.UserID,
			Number:     order.Order,
			Status:     string(order.Status),
			Value:      order.Value,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		}
		list = append(list, *listitem)
	}
	return list, nil
}
