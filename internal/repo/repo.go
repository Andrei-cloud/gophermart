package repo

import (
	"errors"
	"time"
)

var (
	ErrAlreadyExists = errors.New("item already exists")
	ErrNotExists     = errors.New("item not exists")
)

type OrderType string

const (
	CREDIT OrderType = "credit"
	DEBIT  OrderType = "debit"
)

type OrderStatus string

const (
	NEW        OrderStatus = "NEW"
	PROCESSING OrderStatus = "PROCESSING"
	INVALID    OrderStatus = "INVALID"
	PROCESSED  OrderStatus = "PROCESSED"
)

type User struct {
	ID         int64
	Username   string
	Password   string
	Balance    float64
	Withdrawal float64
	CreatedAt  time.Time
}

type Order struct {
	ID         int64
	Order      string
	Type       OrderType
	UserID     int64
	Value      float64
	Status     OrderStatus
	UploadedAt time.Time
}

type Repository interface {
	UserCreate(*User) (int64, error)
	UserGet(string) (*User, error)
	UserGetByID(int64) (*User, error)
	UserDelete(string) error
	UserUpdate(*User) error

	OrderCreate(*Order) (int64, error)
	OrderGet(string) (*Order, error)
	OrderGetList(int64, OrderType) ([]Order, error)
	OrderDelete(string) error
	OrderToProcess() ([]string, error)
	OrderUpdate(string, OrderStatus, float64) error
}
