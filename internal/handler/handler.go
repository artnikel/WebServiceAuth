package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/artnikel/WebServiceAuth/internal/config"
	"github.com/artnikel/WebServiceAuth/internal/model"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gopkg.in/boj/redistore.v1"
)

type UserService interface {
	SignUp(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, user *model.User) (uuid.UUID, error)
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
	cfg      config.Variables
}

func NewEntityUser(srvcUser UserService, srvcBal BalanceService, validate *validator.Validate, cfg config.Variables) *EntityUser {
	return &EntityUser{srvcUser: srvcUser, srvcBal: srvcBal, validate: validate, cfg: cfg}
}

func NewRedisStore(cfg config.Variables) *redistore.RediStore {
	store, err := redistore.NewRediStore(10, "tcp", ":6379", "", []byte(cfg.TokenSignature))
	if err != nil {
		log.Fatalf("Failed to create redis store: %v", err)
	}
	return store
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
	tempPassword := user.Password
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
	user.Password = tempPassword
	userID, err := eu.srvcUser.GetByLogin(c.Request().Context(), &user)
	if err != nil {
		logrus.Errorf("EntityUser-v: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to log in")
	}
	store := NewRedisStore(eu.cfg)
	c.Set("SESSION_ID", uuid.NewString())
	session, _ := store.Get(c.Request(), c.Get("SESSION_ID").(string))
	session.Values["id"] = userID.String()
	session.Values["login"] = user.Login
	session.Values["password"] = user.Password
	session.Values["admin"] = user.Admin
	if err = session.Save(c.Request(), c.Response()); err != nil {
		logrus.Errorf("EntityUser-Login: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "error saving session")
	}
	return c.Redirect(http.StatusSeeOther, "/index")
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
		}).Errorf("Handler-Login: error: %v", err)
		return tmpl.ExecuteTemplate(c.Response().Writer, "auth", map[string]string{
			"errorMsg": "The fields have not been validated",
		})
	}
	userID, err := eu.srvcUser.GetByLogin(c.Request().Context(), &user)
	if err != nil {
		logrus.Errorf("EntityUser-v: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to log in")
	}
	c.Set("SESSION_ID", uuid.NewString())
	store := NewRedisStore(eu.cfg)
	session, _ := store.Get(c.Request(), c.Get("SESSION_ID").(string))
	session.Values["id"] = userID.String()
	session.Values["login"] = user.Login
	session.Values["password"] = user.Password
	session.Values["admin"] = user.Admin
	if err = session.Save(c.Request(), c.Response()); err != nil {
		logrus.Errorf("EntityUser-Login: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "error saving session")
	}
	return c.Redirect(http.StatusSeeOther, "/index")

}

func (eu *EntityUser) GetBalance(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/index/index.html")
	if err != nil {
		return template.ExecError{Name: "index", Err: err}
	}
	store := NewRedisStore(eu.cfg)
	session, _ := store.Get(c.Request(), c.Get("SESSION_ID").(string))
	if len(session.Values) == 0 {
		return c.String(http.StatusOK, "empty result")
	}
	id := session.Values["id"].(string)
	idUUID, err := uuid.Parse(id)
	if err != nil {
		return tmpl.ExecuteTemplate(c.Response().Writer, "index", map[string]string{
			"errorMsg": "Invalid id",
		})
	}
	money, err := eu.srvcBal.GetBalance(c.Request().Context(), idUUID)
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
