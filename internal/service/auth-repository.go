package service

import (
	"context"
	"fmt"

	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	SignUp(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, user *model.User) (string, error)
}

type UserService struct {
	uRep UserRepository
}

func NewUserService(uRep UserRepository) *UserService {
	return &UserService{uRep: uRep}
}

func (us *UserService) SignUp(ctx context.Context, user *model.User) error {
	user.ID = uuid.New()
	user.Admin = false
	var err error
	user.Password, err = us.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("UserService-SignUp-HashPassword: error: %w", err)
	}
	err = us.uRep.SignUp(ctx, user)
	if err != nil {
		return fmt.Errorf("UserService-SignUp: error: %w", err)
	}
	return nil
}

func (us *UserService) GetByLogin(ctx context.Context, user *model.User) (string, error) {
	passwordHash, err := us.uRep.GetByLogin(ctx, user)
	if err != nil {
		return "", fmt.Errorf("UserService-GetByLogin: error: %w", err)
	}
	return passwordHash, nil
}

func (us *UserService) HashPassword(password string) (string, error) {
	const cost = 14
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

func (us *UserService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
