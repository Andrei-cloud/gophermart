package inmem

import (
	"github.com/andrei-cloud/gophermart/internal/repo"
)

type inMemRepo struct {
	userDB      map[string]repo.User
	orderDB     map[string]repo.Order
	nextUserID  int64
	nextOrderID int64
}

func NewInMemRepo() *inMemRepo {
	return &inMemRepo{
		userDB:  make(map[string]repo.User),
		orderDB: make(map[string]repo.Order),
	}
}

func (r *inMemRepo) UserCreate(u *repo.User) (int64, error) {
	if _, ok := r.userDB[u.Username]; ok {
		return 0, repo.ErrAlreadyExists
	}
	u.ID = r.GetNextUserID()
	r.userDB[u.Username] = *u
	return u.ID, nil
}

func (r *inMemRepo) UserGet(name string) (*repo.User, error) {
	var (
		user repo.User
		ok   bool
	)
	if user, ok = r.userDB[name]; !ok {
		return nil, repo.ErrNotExists
	}

	return &user, nil
}

func (r *inMemRepo) UserGetByID(id int64) (*repo.User, error) {
	for _, user := range r.userDB {
		if user.ID == id {
			return &user, nil
		}
	}

	return nil, repo.ErrNotExists
}

func (r *inMemRepo) UserDelete(name string) error {
	var (
		ok bool
	)
	if _, ok = r.userDB[name]; !ok {
		return repo.ErrNotExists
	}

	delete(r.userDB, name)
	return nil
}

func (r *inMemRepo) UserUpdate(u *repo.User) error {
	r.userDB[u.Username] = *u
	return nil
}

func (r *inMemRepo) OrderCreate(o *repo.Order) (int64, error) {
	if _, ok := r.orderDB[o.Order]; ok {
		return 0, repo.ErrAlreadyExists
	}
	o.ID = r.GetNextOrderID()
	r.orderDB[o.Order] = *o
	return o.ID, nil
}

func (r *inMemRepo) OrderGet(name string) (*repo.Order, error) {
	var (
		order repo.Order
		ok    bool
	)
	if order, ok = r.orderDB[name]; !ok {
		return nil, repo.ErrNotExists
	}

	return &order, nil
}

func (r *inMemRepo) OrderGetList(uid int64, t repo.OrderType) ([]repo.Order, error) {
	var (
		orders []repo.Order
	)
	orders = make([]repo.Order, 0)
	for _, order := range r.orderDB {
		if order.UserID == uid && order.Type == t {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

func (r *inMemRepo) OrderDelete(name string) error {
	var (
		ok bool
	)
	if _, ok = r.orderDB[name]; !ok {
		return repo.ErrNotExists
	}

	delete(r.orderDB, name)
	return nil
}

func (r *inMemRepo) GetNextUserID() int64 {
	r.nextUserID++
	return r.nextUserID
}

func (r *inMemRepo) GetNextOrderID() int64 {
	r.nextOrderID++
	return r.nextUserID
}

func (r *inMemRepo) OrderToProcess() ([]string, error) {
	orders := make([]string, 0)
	for _, order := range r.orderDB {
		if order.Status != "PROCESSED" && order.Status != "INVALID" && order.Status != "" {
			orders = append(orders, order.Order)
		}
	}
	return orders, nil
}

func (r *inMemRepo) OrderUpdate(number string, status repo.OrderStatus, accrual float64) error {
	var (
		ok    bool
		order repo.Order
	)
	if order, ok = r.orderDB[number]; !ok {
		return repo.ErrNotExists
	}

	order.Status = status
	order.Value = accrual

	r.orderDB[number] = order
	return nil
}
