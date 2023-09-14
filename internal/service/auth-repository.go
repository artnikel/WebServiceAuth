package service

import (
	"context"
	"fmt"

	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/google/uuid"
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
	err := us.uRep.SignUp(ctx, user)
	if err != nil {
		return fmt.Errorf("UserService-SignUp: error: %w", err)
	}
	return nil
}

func (us *UserService) GetByLogin(ctx context.Context, user *model.User) (string, error) {
	password, err := us.uRep.GetByLogin(ctx, user)
	if err != nil {
		return "", fmt.Errorf("UserService-GetByLogin: error: %w", err)
	}
	return password, nil
}
