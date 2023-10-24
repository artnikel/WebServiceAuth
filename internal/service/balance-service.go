package service

import (
	"context"
	"fmt"

	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BalanceRepository interface {
	BalanceOperation(ctx context.Context, balance *model.Balance) error
	GetBalance(ctx context.Context, profileID uuid.UUID) (float64, error)
}

type BalanceService struct {
	bRep BalanceRepository
}

func NewBalanceService(bRep BalanceRepository) *BalanceService {
	return &BalanceService{bRep: bRep}
}

func (b *BalanceService) BalanceOperation(ctx context.Context, balance *model.Balance) error {
	balance.BalanceID = uuid.New()
	if balance.Operation.IsNegative() {
		money, err := b.GetBalance(ctx, balance.ProfileID)
		if err != nil {
			return fmt.Errorf("getBalance %w", err)
		}
		if decimal.NewFromFloat(money).Cmp(balance.Operation.Abs()) == 1 {
			err = b.bRep.BalanceOperation(ctx, balance)
			if err != nil {
				return fmt.Errorf("balanceOperation %w", err)
			}
			return nil
		}
		return fmt.Errorf("not enough money")
	}
	err := b.bRep.BalanceOperation(ctx, balance)
	if err != nil {
		return fmt.Errorf("balanceOperation %w", err)
	}
	return nil
}

func (b *BalanceService) GetBalance(ctx context.Context, profileID uuid.UUID) (float64, error) {
	money, err := b.bRep.GetBalance(ctx, profileID)
	if err != nil {
		return 0, fmt.Errorf("getBalance %w", err)
	}
	return money, nil
}
