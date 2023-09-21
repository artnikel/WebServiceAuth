package main

import (
	"context"
	"fmt"
	"log"

	"github.com/artnikel/WebServiceAuth/internal/config"
	"github.com/artnikel/WebServiceAuth/internal/handler"
	"github.com/artnikel/WebServiceAuth/internal/repository"
	"github.com/artnikel/WebServiceAuth/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func connectPostgres(connString string) (*pgxpool.Pool, error) {
	cfgPostgres, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	dbpool, err := pgxpool.NewWithConfig(context.Background(), cfgPostgres)
	if err != nil {
		return nil, err
	}
	return dbpool, nil
}

func connectRedis() (*redis.Client, error) {
	cfg, err := config.New()
	if err != nil {
		log.Fatal("Could not parse config: ", err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisWebAddress,
		DB:   0,
	})
	_, err = client.Ping(client.Context()).Result()
	if err != nil {
		return nil, fmt.Errorf("error in method client.Ping(): %v", err)
	}
	return client, nil
}

func main() {
	v := validator.New()
	cfg, err := config.New()
	if err != nil {
		log.Fatal("Could not parse config: ", err)
	}
	dbpool, errPool := connectPostgres(cfg.PostgresConnWebAuth)
	if errPool != nil {
		log.Fatal("could not construct the pool: ", errPool)
	}
	defer dbpool.Close()
	redisClient, err := connectRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer func() {
		errClose := redisClient.Close()
		if errClose != nil {
			log.Fatalf("Failed to disconnect from Redis: %v", errClose)
		}
	}()
	pgRep := repository.NewPgRepository(dbpool)
	userServ := service.NewUserService(pgRep, *cfg)
	balServ := service.NewBalanceService(pgRep)
	hndl := handler.NewEntityUser(userServ, balServ, v, *cfg)
	e := echo.New()
	e.Static("/templates", "templates")
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	store := handler.NewRedisStore(*cfg)
	store.SetMaxAge(10 * 24 * 3600)
	e.Use(session.Middleware(store))
	e.GET("/", hndl.Auth)
	e.POST("/signup", hndl.SignUp)
	e.POST("/login", hndl.Login)
	e.GET("/index", hndl.Index)
	e.POST("/deleteaccount", hndl.DeleteAccount)
	e.GET("/getbalance", hndl.GetBalance)
	e.POST("/deposit", hndl.BalanceOperation)
	e.POST("/buy", hndl.BuyProducts)
	e.Logger.Fatal(e.Start(":8900"))
}
