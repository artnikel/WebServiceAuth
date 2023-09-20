package model

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	ID           uuid.UUID
	Login        string `json:"login" form:"login" validate:"required,min=5,max=20"`
	Password     string `json:"password" form:"password" validate:"required,min=8"`
	Admin        bool
}

type Balance struct {
	BalanceID uuid.UUID
	ProfileID uuid.UUID
	Operation decimal.Decimal `json:"operation" validate:"required"`
}
