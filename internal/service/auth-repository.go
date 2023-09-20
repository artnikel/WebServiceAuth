package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/artnikel/WebServiceAuth/internal/config"
	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	SignUp(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, user *model.User) (string, error)
	AddRefreshToken(ctx context.Context, user *model.User) error
}

type UserService struct {
	uRep UserRepository
	cfg  config.Variables
}

const (
	accessTokenExpiration  = 15 * time.Minute
	refreshTokenExpiration = 72 * time.Hour
	bcryptCost             = 14
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
	tokenPair, err := us.GenerateTokenPair(user.ID, user.Admin)
	if err != nil {
		return fmt.Errorf("UserService-SignUp-GenerateTokenPair: error: %w", err)
	}
	sum := sha256.Sum256([]byte(tokenPair.RefreshToken))
	hashedRefreshToken, err := us.GenerateHash(string(sum[:]))
	if err != nil {
		return fmt.Errorf("UserService-Login-GenerateHash: error: %w", err)
	}
	user.RefreshToken = hashedRefreshToken
	err = us.uRep.SignUp(ctx, user)
	if err != nil {
		return fmt.Errorf("UserService-SignUp: error: %w", err)
	}
	err = us.uRep.AddRefreshToken(ctx, user)
	if err != nil {
		return fmt.Errorf("UserService-AddRefreshToken: error: %w", err)
	}
	return nil
}

func (us *UserService) GetByLogin(ctx context.Context, user *model.User) (*model.TokenPair, error) {
	passwordHash, err := us.uRep.GetByLogin(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("UserService-GetByLogin: error: %w", err)
	}
	if !us.CheckPasswordHash(user.Password, passwordHash) {
		return nil, fmt.Errorf("UserService-GetByLogin: wrong password")
	}
	tokenPair, err := us.GenerateTokenPair(user.ID, user.Admin)
	if err != nil {
		return nil, fmt.Errorf("UserService-SignUp-GenerateTokenPair: error: %w", err)
	}
	err = us.uRep.AddRefreshToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("UserService-AddRefreshToken: error: %w", err)
	}
	return tokenPair, nil
}

func (us *UserService) GenerateTokenPair(id uuid.UUID, admin bool) (*model.TokenPair, error) {
	accessToken, err := us.GenerateJWTToken(accessTokenExpiration, id, admin)
	if err != nil {
		return &model.TokenPair{}, fmt.Errorf("UserService-GenerateTokenPair-GenerateJWTToken: error: %w", err)
	}
	refreshToken, err := us.GenerateJWTToken(refreshTokenExpiration, id, admin)
	if err != nil {
		return &model.TokenPair{}, fmt.Errorf("UserService-GenerateTokenPair-GenerateJWTToken: error: %w", err)
	}
	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (us *UserService) GenerateJWTToken(expiration time.Duration, id uuid.UUID, admin bool) (string, error) {
	claims := &jwt.MapClaims{
		"exp":   time.Now().Add(expiration).Unix(),
		"id":    id,
		"admin": admin,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(us.cfg.TokenSignature))
	if err != nil {
		return "", fmt.Errorf("UserService-GenerateJWTToken: error in method token.SignedString: %w", err)
	}
	return tokenString, nil
}

func (us *UserService) GetIDByToken(authHeader string) (uuid.UUID, error) {
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return uuid.Nil, fmt.Errorf("BalanceService-GetIDByToken: authorization header is invalid")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(us.cfg.TokenSignature), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("BalanceService-GetIDByToken: error jwt parse")
	}
	if !token.Valid {
		return uuid.Nil, fmt.Errorf("BalanceService-GetIDByToken: access token is invalid")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if id, ok := claims["id"].(string); ok {
			profileid, err := uuid.Parse(id)
			if err != nil {
				return uuid.Nil, fmt.Errorf("BalanceService-GetIDByToken: failed to parse")
			}
			return profileid, nil
		}
	}
	return uuid.Nil, nil
}

func (us *UserService) GenerateHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

func (us *UserService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
