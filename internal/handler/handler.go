package handler

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strconv"

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
	SignUpAdmin(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, user *model.User) (uuid.UUID, bool, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
}

type BalanceService interface {
	BalanceOperation(ctx context.Context, balance *model.Balance) error
	GetBalance(ctx context.Context, profileID uuid.UUID) (float64, error)
}

type CartService interface {
	AddCartItems(ctx context.Context, profileid string, carts []model.CartItem) error
	ShowCart(ctx context.Context, profileid string) ([]model.CartItem, error)
}

type EntityUser struct {
	srvcUser UserService
	srvcBal  BalanceService
	srvcCart CartService
	validate *validator.Validate
	cfg      config.Variables
}

func NewEntityUser(srvcUser UserService, srvcBal BalanceService, srvcCart CartService, validate *validator.Validate, cfg config.Variables) *EntityUser {
	return &EntityUser{srvcUser: srvcUser, srvcBal: srvcBal, srvcCart: srvcCart, validate: validate, cfg: cfg}
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
		return echo.ErrNotFound
	}
	tmpl.ExecuteTemplate(c.Response().Writer, "auth", nil)
	return nil
}

func (eu *EntityUser) Index(c echo.Context) error {
	type PageData struct {
		Items []model.CartItem
	}
	tmpl, err := template.ParseFiles("templates/index/index.html")
	if err != nil {
		return echo.ErrNotFound
	}
	cookie, err := c.Cookie("SESSION_ID")
	if err != nil {
		logrus.Errorf("EntityUser-Index: err:%v", err)
		return c.Redirect(http.StatusSeeOther, "/")
	}
	store := NewRedisStore(eu.cfg)
	session, err := store.Get(c.Request(), cookie.Name)
	if err != nil {
		logrus.Errorf("EntityUser-Index: err:%v", err)
		return echo.ErrNotFound
	}
	if len(session.Values) == 0 {
		return c.Redirect(http.StatusSeeOther, "/")
	}
	profileid := session.Values["id"].(string)
	admin, ok := session.Values["admin"].(bool)
	if !ok {
		return tmpl.ExecuteTemplate(c.Response().Writer, "index", nil)
	}
	items, err := eu.srvcCart.ShowCart(c.Request().Context(), profileid)
	if err != nil {
		logrus.Errorf("EntityUser-ShowCart: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to ShowCart cart")
	}
	//ItemsData := PageData{Items: items}
	_, ok = session.Values["balance"].(float64)
	if !ok {
		return tmpl.ExecuteTemplate(c.Response().Writer, "index", struct {
			IsAdmin bool
			Balance float64
			ItemsData PageData
			
		}{
			IsAdmin: admin,
			Balance: 0.0,
			ItemsData: PageData{Items: items},
		})
	}
	return tmpl.ExecuteTemplate(c.Response().Writer, "index", struct {
		IsAdmin bool
		Balance float64
		ItemsData PageData
	}{
		IsAdmin: admin,
		Balance: session.Values["balance"].(float64),
		ItemsData: PageData{Items: items},
	})
}

func (eu *EntityUser) SignUp(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/auth/auth.html")
	if err != nil {
		return echo.ErrNotFound
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
	userID, isAdmin, err := eu.srvcUser.GetByLogin(c.Request().Context(), &user)
	if err != nil {
		logrus.Errorf("EntityUser-SignUp-GetByLogin: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to log in")
	}
	store := NewRedisStore(eu.cfg)
	session, err := store.Get(c.Request(), "SESSION_ID")
	if err != nil {
		logrus.Errorf("EntityUser-SignUp: err:%v", err)
		return echo.ErrNotFound
	}
	session.Values["id"] = userID.String()
	session.Values["login"] = user.Login
	session.Values["password"] = user.Password
	session.Values["admin"] = isAdmin
	session.Values["balance"] = 0.0
	if err = session.Save(c.Request(), c.Response()); err != nil {
		logrus.Errorf("EntityUser-SignUp: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "error saving session")
	}
	return c.Redirect(http.StatusSeeOther, "/index")
}

func (eu *EntityUser) Login(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/auth/auth.html")
	if err != nil {
		return echo.ErrNotFound
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
		}).Errorf("EntityUser-Login: error: %v", err)
		return tmpl.ExecuteTemplate(c.Response().Writer, "auth", map[string]string{
			"errorMsg": "The fields have not been validated",
		})
	}
	userID, isAdmin, err := eu.srvcUser.GetByLogin(c.Request().Context(), &user)
	if err != nil {
		logrus.Errorf("EntityUser-Login: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to log in")
	}
	store := NewRedisStore(eu.cfg)
	session, err := store.Get(c.Request(), "SESSION_ID")
	if err != nil {
		logrus.Errorf("EntityUser-Login: err:%v", err)
		return echo.ErrNotFound
	}
	session.Values["id"] = userID.String()
	session.Values["login"] = user.Login
	session.Values["password"] = user.Password
	session.Values["admin"] = isAdmin
	if err = session.Save(c.Request(), c.Response().Writer); err != nil {
		logrus.Errorf("EntityUser-Login: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "error saving session")
	}
	return c.Redirect(http.StatusSeeOther, "/index")
}

func (eu *EntityUser) DeleteAccount(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/index/index.html")
	if err != nil {
		return echo.ErrNotFound
	}
	store := NewRedisStore(eu.cfg)
	cookie, err := c.Cookie("SESSION_ID")
	if err != nil {
		logrus.Errorf("EntityUser-DeleteAccount: err:%v", err)
		return echo.ErrUnauthorized
	}
	session, err := store.Get(c.Request(), cookie.Name)
	if err != nil {
		logrus.Errorf("EntityUser-DeleteAccount: err:%v", err)
		return echo.ErrNotFound
	}
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
	err = eu.srvcUser.DeleteAccount(c.Request().Context(), idUUID)
	if err != nil {
		logrus.Errorf("EntityUser-DeleteAccount: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete account")
	}
	session.Options.MaxAge = -1
	if err = session.Save(c.Request(), c.Response().Writer); err != nil {
		logrus.Errorf("EntityUser-DeleteAccount: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "error saving session")
	}
	return c.Redirect(http.StatusSeeOther, "/")
}

func (eu *EntityUser) GetBalance(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/index/index.html")
	if err != nil {
		return echo.ErrNotFound
	}
	store := NewRedisStore(eu.cfg)
	cookie, err := c.Cookie("SESSION_ID")
	if err != nil {
		logrus.Errorf("EntityUser-GetBalance: err:%v", err)
		return echo.ErrUnauthorized
	}
	session, err := store.Get(c.Request(), cookie.Name)
	if err != nil {
		logrus.Errorf("EntityUser-GetBalance: err:%v", err)
		return echo.ErrNotFound
	}
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
	session.Values["balance"] = money
	if err = session.Save(c.Request(), c.Response().Writer); err != nil {
		logrus.Errorf("EntityUser-GetBalance: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "error saving session")
	}
	return c.Redirect(http.StatusSeeOther, "/index")
}

func (eu *EntityUser) BalanceOperation(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/index/index.html")
	if err != nil {
		return echo.ErrNotFound
	}
	store := NewRedisStore(eu.cfg)
	cookie, err := c.Cookie("SESSION_ID")
	if err != nil {
		logrus.Errorf("EntityUser-BalanceOperation: err:%v", err)
		return echo.ErrUnauthorized
	}
	session, err := store.Get(c.Request(), cookie.Name)
	if err != nil {
		logrus.Errorf("EntityUser-BalanceOperation: err:%v", err)
		return echo.ErrNotFound
	}
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
	money, err := strconv.ParseFloat(c.FormValue("money"), 64)
	if err != nil {
		return tmpl.ExecuteTemplate(c.Response().Writer, "index", map[string]string{
			"errorMsg": "Invalid sum of money",
		})
	}
	balance := &model.Balance{
		ProfileID: idUUID,
		Operation: decimal.NewFromFloat(money),
	}
	err = eu.srvcBal.BalanceOperation(c.Request().Context(), balance)
	if err != nil {
		logrus.Errorf("EntityUser-BalanceOperation: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to made balance operation")
	}
	return c.Redirect(http.StatusSeeOther, "/index")
}

func (eu *EntityUser) BuyProducts(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/index/index.html")
	if err != nil {
		return echo.ErrNotFound
	}
	store := NewRedisStore(eu.cfg)
	cookie, err := c.Cookie("SESSION_ID")
	if err != nil {
		logrus.Errorf("EntityUser-BuyProducts: err:%v", err)
		return echo.ErrUnauthorized
	}
	session, err := store.Get(c.Request(), cookie.Name)
	if err != nil {
		logrus.Errorf("EntityUser-BuyProducts: err:%v", err)
		return echo.ErrNotFound
	}
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
	var data struct {
		TotalPrice float64 `json:"totalPrice"`
	}
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}
	money, err := eu.srvcBal.GetBalance(c.Request().Context(), idUUID)
	if err != nil {
		logrus.Errorf("EntityUser-BuyProducts-GetBalance: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get balance")
	}
	if decimal.NewFromFloat(money).Cmp(decimal.NewFromFloat(data.TotalPrice)) == -1 {
		return c.JSON(http.StatusBadRequest, "Not enough money")
	}
	balance := &model.Balance{
		ProfileID: idUUID,
		Operation: decimal.NewFromFloat(data.TotalPrice).Neg(),
	}
	err = eu.srvcBal.BalanceOperation(c.Request().Context(), balance)
	if err != nil {
		logrus.Errorf("EntityUser-GetBalance: err:%v", err)
		return c.Redirect(http.StatusSeeOther, "/index")
	}
	return c.Redirect(http.StatusSeeOther, "/index")
}

func (eu *EntityUser) SignUpAdmin(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/index/index.html")
	if err != nil {
		return echo.ErrNotFound
	}
	var user model.User
	user.Login = c.FormValue("login")
	user.Password = c.FormValue("password")
	err = eu.validate.StructCtx(c.Request().Context(), user)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Login":    user.Login,
			"Password": user.Password,
		}).Errorf("Handler-SignUpAdmin: error: %v", err)
		return tmpl.ExecuteTemplate(c.Response().Writer, "auth", map[string]string{
			"errorMsg": "The fields have not been validated",
		})
	}
	err = eu.srvcUser.SignUpAdmin(c.Request().Context(), &user)
	if err != nil {
		logrus.Errorf("EntityUser-SignUpAdmin: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to sign up admin")
	}
	return c.Redirect(http.StatusSeeOther, "/index")
}

func (eu *EntityUser) DeleteByID(c echo.Context) error {
	tmpl, err := template.ParseFiles("templates/index/index.html")
	if err != nil {
		return echo.ErrNotFound
	}
	store := NewRedisStore(eu.cfg)
	cookie, err := c.Cookie("SESSION_ID")
	if err != nil {
		logrus.Errorf("EntityUser-DeleteByID: err:%v", err)
		return echo.ErrUnauthorized
	}
	session, err := store.Get(c.Request(), cookie.Name)
	if err != nil {
		logrus.Errorf("EntityUser-DeleteByID: err:%v", err)
		return echo.ErrNotFound
	}
	if len(session.Values) == 0 {
		return c.String(http.StatusOK, "empty result")
	}
	if !session.Values["admin"].(bool) {
		return c.String(http.StatusMethodNotAllowed, "you`re not admin")
	}
	id, err := uuid.Parse(c.FormValue("profileid"))
	if err != nil {
		return tmpl.ExecuteTemplate(c.Response().Writer, "index", map[string]string{
			"errorMsg": "Invalid id",
		})
	}
	err = eu.srvcUser.DeleteAccount(c.Request().Context(), id)
	if err != nil {
		logrus.Errorf("EntityUser-DeleteByID-DeleteAccount: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete another account")
	}
	return c.Redirect(http.StatusSeeOther, "/index")
}

func (eu *EntityUser) Logout(c echo.Context) error {
	store := NewRedisStore(eu.cfg)
	cookie, err := c.Cookie("SESSION_ID")
	if err != nil {
		logrus.Errorf("EntityUser-Logout: err:%v", err)
		return echo.ErrUnauthorized
	}
	session, err := store.Get(c.Request(), cookie.Name)
	if err != nil {
		logrus.Errorf("EntityUser-Logout: err:%v", err)
		return echo.ErrNotFound
	}
	session.Options.MaxAge = -1
	session.Save(c.Request(), c.Response().Writer)
	return c.Redirect(http.StatusSeeOther, "/")
}

func (eu *EntityUser) SaveCart(c echo.Context) error {
	var cartItems []model.CartItem
	if err := c.Bind(&cartItems); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	store := NewRedisStore(eu.cfg)
	cookie, err := c.Cookie("SESSION_ID")
	if err != nil {
		logrus.Errorf("EntityUser-SaveCart: err:%v", err)
		return echo.ErrUnauthorized
	}
	session, err := store.Get(c.Request(), cookie.Name)
	if err != nil {
		logrus.Errorf("EntityUser-SaveCart: err:%v", err)
		return echo.ErrNotFound
	}
	if len(session.Values) == 0 {
		return c.String(http.StatusOK, "empty result")
	}
	profileid := session.Values["id"].(string)

	err = eu.srvcCart.AddCartItems(c.Request().Context(), profileid, cartItems)
	if err != nil {
		logrus.Errorf("EntityUser-SaveCart: err:%v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save cart")
	}
	return c.Redirect(http.StatusSeeOther, "/index")
}

// func (eu *EntityUser) ShowCart(c echo.Context) error {
// 	type PageData struct {
// 		Items []model.CartItem
// 	}
// 	tmpl, err := template.ParseFiles("templates/index/index.html")
// 	if err != nil {
// 		return echo.ErrNotFound
// 	}
// 	store := NewRedisStore(eu.cfg)
// 	cookie, err := c.Cookie("SESSION_ID")
// 	if err != nil {
// 		logrus.Errorf("EntityUser-SaveCart: err:%v", err)
// 		return echo.ErrUnauthorized
// 	}
// 	session, err := store.Get(c.Request(), cookie.Name)
// 	if err != nil {
// 		logrus.Errorf("EntityUser-SaveCart: err:%v", err)
// 		return echo.ErrNotFound
// 	}
// 	if len(session.Values) == 0 {
// 		return c.String(http.StatusOK, "empty result")
// 	}
// 	profileid := session.Values["id"].(string)
// 	items, err := eu.srvcCart.ShowCart(c.Request().Context(), profileid)
// 	if err != nil {
// 		logrus.Errorf("EntityUser-ShowCart: err:%v", err)
// 		return echo.NewHTTPError(http.StatusInternalServerError, "failed to ShowCart cart")
// 	}
// 	data := PageData{Items: items}
// 	return tmpl.ExecuteTemplate(c.Response().Writer, "index", data)
// }
