package service

import (
	"context"
	"fmt"

	"github.com/artnikel/WebServiceAuth/internal/model"
)

type CartRepository interface {
	Set(ctx context.Context, profileid string, cart []model.CartItem) error
	Get(ctx context.Context, profileid string) (carts []model.CartItem, e error)
	Delete(ctx context.Context, profileid string) error
}

type CartService struct {
	cartRep CartRepository
}

func NewCartService(cartRep CartRepository) *CartService {
	return &CartService{cartRep: cartRep}
}

func (cs *CartService) AddCartItems(ctx context.Context, profileid string, carts []model.CartItem) error {
	err := cs.cartRep.Set(ctx, profileid, carts)
	if err != nil {
		return fmt.Errorf("addCartItems %w", err)
	}
	return nil
}

func (cs *CartService) ShowCart(ctx context.Context, profileid string) ([]model.CartItem, error) {
	carts, err := cs.cartRep.Get(ctx, profileid)
	if err != nil {
		return nil, fmt.Errorf("get %w", err)
	}
	return carts, nil
}

func (cs *CartService) DeleteCart(ctx context.Context, profileid string) error {
	err := cs.cartRep.Delete(ctx, profileid)
	if err != nil {
		return fmt.Errorf("delete %w", err)
	}
	return nil
}
