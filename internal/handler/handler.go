package handler

import (
	"context"
	"net/http"
	"text/template"

	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type UserService interface {
	SignUp(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, user *model.User) (string, error)
}

type EntityUser struct {
	srvcUser UserService
	validate *validator.Validate
}

func NewEntityUser(srvcUser UserService, validate *validator.Validate) *EntityUser {
	return &EntityUser{srvcUser: srvcUser, validate: validate}
}

func (eu *EntityUser) Auth(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/auth/auth.html")
	if err != nil {
		return template.ExecError{Name: "auth", Err: err}
	}
	tmpl.ExecuteTemplate(c.Response().Writer, "auth", nil)
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
	http.Redirect(c.Response().Writer, c.Request(), "/catalog", http.StatusSeeOther)
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
	password, err := eu.srvcUser.GetByLogin(c.Request().Context(), &user)
	if err != nil {
		logrus.Errorf("EntityUser-SignUp: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to log in")
	}
	if user.Password == password {
		http.Redirect(c.Response().Writer, c.Request(), "/catalog", http.StatusSeeOther)
		return nil
	}
	return tmpl.ExecuteTemplate(c.Response().Writer, "auth", map[string]string{
		"errorMsg": "Wrong password",
	})
}
