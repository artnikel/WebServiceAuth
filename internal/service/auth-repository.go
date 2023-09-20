package service

import (
	"context"
	"fmt"

	"github.com/artnikel/WebServiceAuth/internal/config"
	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	SignUp(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, user *model.User) (uuid.UUID, string, error)
}

type UserService struct {
	uRep UserRepository
	cfg  config.Variables
}

const (
	bcryptCost = 14
)

func NewUserService(uRep UserRepository, cfg config.Variables) *UserService {
	return &UserService{uRep: uRep, cfg: cfg}
}

func (us *UserService) SignUp(ctx context.Context, user *model.User) error {
	user.ID = uuid.New()
	user.Admin = false
	var err error
	user.Password, err = us.GenerateHash(user.Password)
	if err != nil {
		return fmt.Errorf("UserService-SignUp-GenerateHash: error: %w", err)
	}
	err = us.uRep.SignUp(ctx, user)
	if err != nil {
		return fmt.Errorf("UserService-SignUp: error: %w", err)
	}
	return nil
}

func (us *UserService) GetByLogin(ctx context.Context, user *model.User) (uuid.UUID, error) {
	userID, passwordHash, err := us.uRep.GetByLogin(ctx, user)
	if err != nil {
		return uuid.Nil, fmt.Errorf("UserService-GetByLogin: error: %w", err)
	}
	if !us.CheckPasswordHash(user.Password, passwordHash) {
		return uuid.Nil, fmt.Errorf("UserService-GetByLogin: wrong password")
	}
	return userID, nil
}

func (us *UserService) GenerateHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

func (us *UserService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
