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
	GetByLogin(ctx context.Context, user *model.User) (uuid.UUID, string, bool, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
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
	var err error
	user.Password, err = us.GenerateHash(user.Password)
	if err != nil {
		return fmt.Errorf("generateHash %w", err)
	}
	err = us.uRep.SignUp(ctx, user)
	if err != nil {
		return fmt.Errorf("signUp %w", err)
	}
	return nil
}

func (us *UserService) SignUpAdmin(ctx context.Context, user *model.User) error {
	user.Admin = true
	err := us.SignUp(ctx, user)
	if err != nil {
		return fmt.Errorf("signUp %w", err)
	}
	return nil
}

func (us *UserService) GetByLogin(ctx context.Context, user *model.User) (uuid.UUID, bool, error) {
	userID, passwordHash, admin, err := us.uRep.GetByLogin(ctx, user)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("getByLogin %w", err)
	}
	if !us.CheckPasswordHash(user.Password, passwordHash) {
		return uuid.Nil, false, fmt.Errorf("wrong password")
	}
	return userID, admin, nil
}

func (us *UserService) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	err := us.uRep.DeleteAccount(ctx, id)
	if err != nil {
		return fmt.Errorf("deleteAccount %w", err)
	}
	return nil
}

func (us *UserService) GenerateHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

func (us *UserService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
