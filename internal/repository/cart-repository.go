package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/go-redis/redis/v8"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRepositoryRedis(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func (rds *RedisRepository) Set(ctx context.Context, profileid string, carts []model.CartItem) error {
	if carts == nil {
		return nil
	}
	cartJSON, err := json.Marshal(&carts)
	if err != nil {
		return fmt.Errorf("marshal %w", err)
	}
	err = rds.client.HSet(ctx, "carts", profileid, cartJSON).Err()
	if err != nil {
		return fmt.Errorf("hSet %w", err)
	}
	return nil
}

func (rds *RedisRepository) Get(ctx context.Context, profileid string) (carts []model.CartItem, e error) {
	cartJSON, err := rds.client.HGet(ctx, "carts", profileid).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("hGet %w", err)
	}
	err = json.Unmarshal([]byte(cartJSON), &carts)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %w", err)
	}
	return carts, nil
}

func (rds *RedisRepository) Delete(ctx context.Context, profileid string) error {
	_, err := rds.client.HDel(ctx, "carts", profileid).Result()
	if err != nil {
		return fmt.Errorf("hDel %w", err)
	}
	return nil
}
