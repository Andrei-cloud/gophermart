package domain

import (
	"fmt"

	"github.com/andrei-cloud/gophermart/internal/repo"
	"golang.org/x/crypto/bcrypt"
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
	err = u.hashPassword()
	if err != nil {
		return 0, err
	}
	id, err := r.UserCreate(&repo.User{Username: u.Username, Password: u.Password})
	if err != nil {
		return 0, fmt.Errorf("register user failed: %w", err)
	}
	return id, nil
}

func (u *UserModel) Login(r repo.Repository) (int64, error) {
	user, err := r.UserGet(u.Username)
	if err != nil {
		return 0, repo.ErrNotExists
	}

	if !u.checkPasswordHash(user.Password) {
		return 0, repo.ErrNotExists
	}

	return user.ID, nil
}

func (u *UserModel) GetBalance(r repo.Repository) (map[string]float64, error) {
	user, err := r.UserGetByID(u.ID)
	if err != nil {
		return nil, fmt.Errorf("get by id user failed: %w", err)
	}

	return map[string]float64{"current": user.Balance, "withdrawn": user.Withdrawal}, nil
}

func (u *UserModel) hashPassword() error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(u.Password), 8)
	u.Password = string(bytes)
	return err
}

func (u *UserModel) checkPasswordHash(hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(u.Password))
	return err == nil
}
