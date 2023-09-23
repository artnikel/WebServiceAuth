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
	cartJSON, err := json.Marshal(&carts)
	if err != nil {
		return fmt.Errorf("RedisRepository-Set-json.Marshal: error: %w", err)
	}
	_, err = rds.client.HGet(ctx, "carts", profileid).Result()
	if err != nil {
		if err == redis.Nil {
			rds.client.HSet(ctx, "carts", profileid, cartJSON)
			return nil
		}
		return fmt.Errorf("RedisRepository-Set-HGet: error: %w", err)
	}
	return nil
}

func (rds *RedisRepository) Get(ctx context.Context, profileid string) (carts []model.CartItem, e error) {
	cartJSON, err := rds.client.HGet(ctx, "carts", profileid).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("RedisRepository-Get-HGet: error: %w", err)
	}
	err = json.Unmarshal([]byte(cartJSON), &carts)
	if err != nil {
		return nil, fmt.Errorf("RedisRepository-Get-json.Unmarshal: error: %w", err)
	}
	return carts, nil
}

func (rds *RedisRepository) Delete(ctx context.Context, profileid string) error {
	_, err := rds.client.HDel(ctx, "carts", profileid).Result()
	if err != nil {
		return fmt.Errorf("RedisRepository-Delete-HDel: error: %w", err)
	}
	return nil
}
