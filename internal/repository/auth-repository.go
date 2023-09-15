package repository

import (
	"context"
	"fmt"

	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgRepository struct {
	pool *pgxpool.Pool
}

func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{
		pool: pool,
	}
}

func (p *PgRepository) SignUp(ctx context.Context, user *model.User) error {
	var count int
	err := p.pool.QueryRow(ctx, "SELECT COUNT(id) FROM users WHERE login = $1", user.Login).Scan(&count)
	if err != nil {
		return fmt.Errorf("PgRepository-SignUpUser: error in method r.pool.QuerryRow(): %w", err)
	}
	if count != 0 {
		return fmt.Errorf("PgRepository-SignUpUser: the login is occupied by another user")
	}
	_, err = p.pool.Exec(ctx, "INSERT INTO users (id, login, password, admin) VALUES ($1, $2, $3, $4)", user.ID, user.Login, user.Password, user.Admin)
	if err != nil {
		return fmt.Errorf("PgRepository-SignUpUser: error in method r.pool.Exec(): %w", err)
	}
	return nil
}

func (p *PgRepository) GetByLogin(ctx context.Context,user *model.User) (string, error) {
	var count int
	err := p.pool.QueryRow(ctx, "SELECT COUNT(id) FROM users WHERE login = $1", user.Login).Scan(&count)
	if err != nil {
		return "", fmt.Errorf("PgRepository-GetByLogin: error in method r.pool.QuerryRow(): %w", err)
	}
	if count == 0 {
		return "", fmt.Errorf("PgRepository-GetByLogin: the login does not exist ")
	}
	var passwordHash string
	err = p.pool.QueryRow(ctx, "SELECT password FROM users WHERE login = $1", user.Login).Scan(&passwordHash)
	if err != nil {
		return "", fmt.Errorf("PgRepository-GetByLogin: error in method r.pool.QuerryRow(): %w", err)
	}
	return passwordHash, nil
}


