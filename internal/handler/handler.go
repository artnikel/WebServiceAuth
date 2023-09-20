package handler

import (
	"context"
	"net/http"
	"strconv"
	"text/template"

	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type UserService interface {
	SignUp(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, user *model.User) (*model.TokenPair, error)
	CheckPasswordHash(password, hash string) bool
}

type BalanceService interface {
	BalanceOperation(ctx context.Context, balance *model.Balance) error
	GetBalance(ctx context.Context, profileID uuid.UUID) (float64, error)
}

type EntityUser struct {
	srvcUser UserService
	srvcBal  BalanceService
	validate *validator.Validate
}

func NewEntityUser(srvcUser UserService, srvcBal BalanceService, validate *validator.Validate) *EntityUser {
	return &EntityUser{srvcUser: srvcUser, srvcBal: srvcBal, validate: validate}
}

func (eu *EntityUser) Auth(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/auth/auth.html")
	if err != nil {
		return template.ExecError{Name: "auth", Err: err}
	}
	tmpl.ExecuteTemplate(c.Response().Writer, "auth", nil)
	return nil
}

func (eu *EntityUser) Index(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/index/index.html")
	if err != nil {
		return template.ExecError{Name: "index", Err: err}
	}
	tmpl.ExecuteTemplate(c.Response().Writer, "index", nil)
	return nil
}

func (eu *EntityUser) SignUp(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/auth/auth.html")
	if err != nil {
		return template.ExecError{Name: "auth", Err: err}
	}
	var user model.User
	if err := c.Bind(&user); err != nil {
		return tmpl.ExecuteTemplate(c.Response().Writer, "auth", map[string]string{
			"errorMsg": "Failed to bind fields",
		})
	}
	err = eu.validate.StructCtx(c.Request().Context(), user)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Login":    user.Login,
			"Password": user.Password,
		}).Errorf("Handler-SignUp: error: %v", err)
		return tmpl.ExecuteTemplate(c.Response().Writer, "auth", map[string]string{
			"errorMsg": "The fields have not been validated",
		})
	}
	err = eu.srvcUser.SignUp(c.Request().Context(), &user)
	if err != nil {
		logrus.Errorf("EntityUser-SignUp: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to sign up")
	}
	user.Password = c.FormValue("password")
	tokenPair, err := eu.srvcUser.GetByLogin(c.Request().Context(), &user)
	if err != nil {
		logrus.Errorf("EntityUser-SignUp: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to log in")
	}
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		HttpOnly: true,
	})
	return c.JSON(http.StatusOK, map[string]string{
		"access_token": tokenPair.AccessToken,
	})
}

func (eu *EntityUser) Login(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/auth/auth.html")
	if err != nil {
		return template.ExecError{Name: "auth", Err: err}
	}
	var user model.User
	if err := c.Bind(&user); err != nil {
		return tmpl.ExecuteTemplate(c.Response().Writer, "auth", map[string]string{
			"errorMsg": "Failed to bind fields",
		})
	}
	err = eu.validate.StructCtx(c.Request().Context(), user)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Login":    user.Login,
			"Password": user.Password,
		}).Errorf("Handler-SignUp: error: %v", err)
		return tmpl.ExecuteTemplate(c.Response().Writer, "auth", map[string]string{
			"errorMsg": "The fields have not been validated",
		})
	}
	tokenPair, err := eu.srvcUser.GetByLogin(c.Request().Context(), &user)
	if err != nil {
		logrus.Errorf("EntityUser-SignUp: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to log in")
	}
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		HttpOnly: true,
	})
	return c.JSON(http.StatusOK, map[string]string{
		"access_token": tokenPair.AccessToken,
	})
}

func (eu *EntityUser) GetBalance(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/index/index.html")
	if err != nil {
		return template.ExecError{Name: "index", Err: err}
	}
	id, err := uuid.Parse(c.FormValue("profileid"))
	if err != nil {
		return tmpl.ExecuteTemplate(c.Response().Writer, "index", map[string]string{
			"errorMsg": "Invalid id",
		})
	}
	money, err := eu.srvcBal.GetBalance(c.Request().Context(), id)
	if err != nil {
		logrus.Errorf("EntityUser-GetBalance: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get balance")
	}
	tmpl.ExecuteTemplate(c.Response().Writer, "index", money)
	return nil
}

func (eu *EntityUser) BalanceOperation(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/index/index.html")
	if err != nil {
		return template.ExecError{Name: "index", Err: err}
	}
	id, err := uuid.Parse(c.FormValue("profileid"))
	if err != nil {
		return tmpl.ExecuteTemplate(c.Response().Writer, "index", map[string]string{
			"errorMsg": "Invalid id",
		})
	}
	money, err := strconv.ParseFloat(c.FormValue("money"), 64)
	if err != nil {
		return tmpl.ExecuteTemplate(c.Response().Writer, "index", map[string]string{
			"errorMsg": "Invalid sum of money",
		})
	}
	balance := &model.Balance{
		ProfileID: id,
		Operation: decimal.NewFromFloat(money),
	}
	err = eu.srvcBal.BalanceOperation(c.Request().Context(), balance)
	if err != nil {
		logrus.Errorf("EntityUser-GetBalance: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to made balance operation")
	}
	tmpl.ExecuteTemplate(c.Response().Writer, "index", nil)
	return nil
}
