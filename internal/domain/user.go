package domain

import (
	"github.com/andrei-cloud/gophermart/internal/repo"
	"github.com/andrei-cloud/gophermart/pkg/utils"
	"github.com/pkg/errors"
)

type User interface {
	Register(repo.Repository) (int64, error)
	Login(repo.Repository) error
	GetBalance(repo.Repository) (float64, error)
}

type UserModel struct {
	ID       int64  `json:"-"`
	Username string `json:"login"`
	Password string `json:"password"`
}

func (u *UserModel) Register(r repo.Repository) (int64, error) {
	var err error
	u.Password, err = utils.HashPassword(u.Password)
	if err != nil {
		return 0, err
	}
	id, err := r.UserCreate(&repo.User{Username: u.Username, Password: u.Password})
	if err != nil {
		return 0, errors.Wrap(err, "register user failed")
	}
	return id, nil
}

func (u *UserModel) Login(r repo.Repository) (int64, error) {
	user, err := r.UserGet(u.Username)
	if err != nil {
		return 0, repo.ErrNotExists
	}

	if !utils.CheckPasswordHash(u.Password, user.Password) {
		return 0, repo.ErrNotExists
	}

	return user.ID, nil
}

func (u *UserModel) GetBalance(r repo.Repository) (map[string]float64, error) {
	user, err := r.UserGetByID(u.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get by id user failed")
	}

	return map[string]float64{"current": user.Balance, "withdrawn": user.Withdrawal}, nil
}
