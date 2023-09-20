package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/artnikel/WebServiceAuth/internal/config"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	cfg, err := config.New()
	if err != nil {
		log.Fatal("JWTMiddleware: Could not parse config: ", err)
	}
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "JWTMiddleware: Missing authorization header")
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return echo.NewHTTPError(http.StatusUnauthorized, "JWTMiddleware: Invalid authorization header")
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.TokenSignature), nil
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "JWTMiddleware: Invalid token")
		}
		if !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "JWTMiddleware: Invalid token")
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			exp := claims["exp"].(float64)
			if exp < float64(time.Now().Unix()) {
				return echo.NewHTTPError(http.StatusUnauthorized, "JWTMiddleware: Token expired")
			}
		}
		c.Set("users", token)
		return next(c)
	}
}

func JWTMiddlewareAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	cfg, err := config.New()
	if err != nil {
		log.Fatal("JWTMiddleware: Could not parse config: ", err)
	}
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "JWTMiddleware: Missing authorization header")
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return echo.NewHTTPError(http.StatusUnauthorized, "JWTMiddleware: Invalid authorization header")
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.TokenSignature), nil
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "JWTMiddleware: Invalid token")
		}
		if !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "JWTMiddleware: Invalid token")
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			exp := claims["exp"].(float64)
			if exp < float64(time.Now().Unix()) {
				return echo.NewHTTPError(http.StatusUnauthorized, "JWTMiddleware: Token expired")
			}
			isAdmin, adminExists := claims["admin"].(bool)
			if !adminExists || !isAdmin {
				return echo.NewHTTPError(http.StatusForbidden, "JWTMiddleware: Access denied, not an admin")
			}

		}
		c.Set("users", token)
		return next(c)
	}
}