package repository

import (
	"context"
	"fmt"

	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/google/uuid"
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
	_, err = p.pool.Exec(ctx, "INSERT INTO users (id, login, password, admin) VALUES ($1, $2, $3, $4)",
		user.ID, user.Login, user.Password, user.Admin)
	if err != nil {
		return fmt.Errorf("PgRepository-SignUpUser: error in method r.pool.Exec(): %w", err)
	}
	return nil
}

func (p *PgRepository) GetByLogin(ctx context.Context, user *model.User) (uuid.UUID, string, bool, error) {
	var count int
	err := p.pool.QueryRow(ctx, "SELECT COUNT(id) FROM users WHERE login = $1", user.Login).Scan(&count)
	if err != nil {
		return uuid.Nil, "", false, fmt.Errorf("PgRepository-GetByLogin: error in method r.pool.QuerryRow(): %w", err)
	}
	if count == 0 {
		return uuid.Nil, "", false, fmt.Errorf("PgRepository-GetByLogin: the login does not exist ")
	}
	var passwordHash string
	var userID uuid.UUID
	var admin bool
	err = p.pool.QueryRow(ctx, "SELECT id, password, admin FROM users WHERE login = $1", user.Login).Scan(&userID, &passwordHash, &admin)
	if err != nil {
		return uuid.Nil, "", false, fmt.Errorf("PgRepository-GetByLogin: error in method r.pool.QuerryRow(): %w", err)
	}
	return userID, passwordHash, admin, nil
}

func (p *PgRepository) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	var count int
	err := p.pool.QueryRow(ctx, "SELECT COUNT(id) FROM users WHERE id = $1", id).Scan(&count)
	if err != nil {
		return fmt.Errorf("PgRepository-DeleteAccount: error in method r.pool.QuerryRow(): %w", err)
	}
	if count == 0 {
		return fmt.Errorf("PgRepository-DeleteAccount: cannot delete non-existent user")
	}
	_, err = p.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("PgRepository-DeleteAccount: error in method r.pool.Exec(): %w", err)
	}
	return nil
}
