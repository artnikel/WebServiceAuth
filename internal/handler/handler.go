package handler

import (
	"context"
	"net/http"
	"text/template"

	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type UserService interface {
	SignUp(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, user *model.User) (string, error)
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
	user.Login = c.FormValue("login")
	user.Password = c.FormValue("password")
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
	http.Redirect(c.Response().Writer, c.Request(), "/index", http.StatusSeeOther)
	return nil
}

func (eu *EntityUser) Login(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/auth/auth.html")
	if err != nil {
		return template.ExecError{Name: "auth", Err: err}
	}
	var user model.User
	user.Login = c.FormValue("login")
	user.Password = c.FormValue("password")
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
	passwordHash, err := eu.srvcUser.GetByLogin(c.Request().Context(), &user)
	if err != nil {
		logrus.Errorf("EntityUser-SignUp: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to log in")
	}
	if eu.srvcUser.CheckPasswordHash(user.Password,passwordHash) {
		http.Redirect(c.Response().Writer, c.Request(), "/index", http.StatusSeeOther)
		return nil
	}
	return tmpl.ExecuteTemplate(c.Response().Writer, "auth", map[string]string{
		"errorMsg": "Wrong password",
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
