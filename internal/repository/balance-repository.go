package repository

import (
	"context"
	"fmt"

	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (p *PgRepository) BalanceOperation(ctx context.Context, balance *model.Balance) error {
	_, err := p.pool.Exec(ctx, "INSERT INTO balance (balanceid, profileid, operation) VALUES ($1, $2, $3)",
		balance.BalanceID, balance.ProfileID, balance.Operation)
	if err != nil {
		return fmt.Errorf("PgRepository-BalanceOperation: error in method tx.Exec(): %w", err)
	}

	return nil
}

func (p *PgRepository) GetBalance(ctx context.Context, profileID uuid.UUID) (float64, error) {
	rows, err := p.pool.Query(ctx, "SELECT operation FROM balance WHERE profileid = $1 FOR UPDATE", profileID)
	if err != nil {
		return 0, fmt.Errorf("PgRepository-GetBalance: error in method tx.QueryRow(): %w", err)
	}
	defer rows.Close()
	var money decimal.Decimal
	for rows.Next() {
		var operation decimal.Decimal
		err := rows.Scan(&operation)
		if err != nil {
			return 0, fmt.Errorf("BalanceService-GetBalance: error in method rows.Scan:%w", err)
		}
		money = money.Add(operation)
	}
	return money.InexactFloat64(), nil
}
