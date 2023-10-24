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
		log.Fatalf("Could not parse config: %v", err)
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
		log.Fatalf("Could not parse config: %v", err)
	}
	dbpool, errPool := connectPostgres(cfg.PostgresConnWebAuth)
	if errPool != nil {
		log.Fatalf("could not construct the pool: %v", errPool)
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
	cartRep := repository.NewRepositoryRedis(redisClient)
	userServ := service.NewUserService(pgRep, *cfg)
	balServ := service.NewBalanceService(pgRep)
	cartServ := service.NewCartService(cartRep)
	hndl := handler.NewEntityUser(userServ, balServ, cartServ, v, *cfg)
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
	e.POST("/signupadmin", hndl.SignUpAdmin)
	e.POST("/deletebyid", hndl.DeleteByID)
	e.POST("/logout", hndl.Logout)
	e.POST("/savecart", hndl.SaveCart)
	e.POST("/clearcart", hndl.ClearCart)
	e.Logger.Fatal(e.Start(":8900"))
}
