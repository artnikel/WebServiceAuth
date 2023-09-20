package main

import (
	"context"
	"log"

	"github.com/artnikel/WebServiceAuth/internal/config"
	"github.com/artnikel/WebServiceAuth/internal/handler"
	"github.com/artnikel/WebServiceAuth/internal/repository"
	"github.com/artnikel/WebServiceAuth/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	custommiddleware "github.com/artnikel/WebServiceAuth/internal/middleware")

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
	pgRep := repository.NewPgRepository(dbpool)
	userServ := service.NewUserService(pgRep, *cfg)
	balServ := service.NewBalanceService(pgRep)
	hndl := handler.NewEntityUser(userServ, balServ, v)
	e := echo.New()
	e.Static("/templates", "templates")
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/", hndl.Auth)
	e.POST("/signup", hndl.SignUp)
	e.POST("/login", hndl.Login)
	e.GET("/index", hndl.Index, custommiddleware.JWTMiddleware)
	e.GET("/getbalance", hndl.GetBalance, custommiddleware.JWTMiddleware)
	e.POST("/deposit", hndl.BalanceOperation, custommiddleware.JWTMiddleware)
	e.Logger.Fatal(e.Start(":8900"))
}
