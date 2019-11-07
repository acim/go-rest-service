package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	abmiddleware "github.com/acim/go-rest-server/pkg/middleware"
	"github.com/acim/go-rest-server/pkg/store"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"go.uber.org/zap"
)

// Auth controller.
type Auth struct {
	users                  store.Users
	jwtauth                *jwtauth.JWTAuth
	authTokenExpiration    time.Duration
	refreshTokenExpiration time.Duration
	logger                 *zap.Logger
}

// NewAuth creates new auth controller.
func NewAuth(users store.Users, jwtauth *jwtauth.JWTAuth, logger *zap.Logger, opts ...AuthOption) *Auth {
	a := &Auth{
		users:                  users,
		jwtauth:                jwtauth,
		authTokenExpiration:    15 * time.Minute,
		refreshTokenExpiration: 7 * 24 * time.Hour,
		logger:                 logger,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// Login handles /auth/login endpoint.
func (c *Auth) Login(w http.ResponseWriter, r *http.Request) {
	res := abmiddleware.ResponseFromContext(r.Context())

	l := &login{}

	err := json.NewDecoder(r.Body).Decode(l)
	if err != nil {
		c.logger.Warn("login", zap.NamedError("json decode", err))
		res.SetStatusBadRequest(errParsingRequestBody)

		return
	}

	u, err := c.users.FindByEmail(r.Context(), l.Email)
	if err != nil {
		c.logger.Warn("login", zap.NamedError("find by email", err))
		res.SetStatusInternalServerError("")

		return
	}

	if !u.IsValidPassword(l.Password) {
		res.SetStatusForbidden(errInvalidCredentials)
		return
	}

	authToken, err := c.token(c.authTokenExpiration, middleware.GetReqID(r.Context()), u.ID)
	if err != nil {
		c.logger.Warn("login", zap.NamedError("auth token", err))
		res.SetStatusInternalServerError("")

		return
	}

	refreshToken, err := c.token(c.refreshTokenExpiration, middleware.GetReqID(r.Context()), u.ID)
	if err != nil {
		c.logger.Warn("login", zap.NamedError("refresh token", err))
		res.SetStatusInternalServerError("")

		return
	}

	res.SetPayload(&token{AuthToken: authToken, RefreshToken: refreshToken})
}

// User handles /auth/user endpoint.
func (c *Auth) User(w http.ResponseWriter, r *http.Request) {
	res := abmiddleware.ResponseFromContext(r.Context())

	userID, err := getUserID(r.Context())
	if err != nil {
		c.logger.Warn("user", zap.NamedError("get user id", err))
		res.SetStatusInternalServerError(err.Error())
		return
	}

	user, err := c.users.FindByID(r.Context(), userID)
	if err != nil {
		c.logger.Warn("user", zap.NamedError("find by id", err))
		res.SetStatusNotFound(fmt.Sprintf("user %s not found", userID))
		return
	}

	user.Password = ""
	res.SetPayload(user)
}

func (c *Auth) token(expiration time.Duration, requestID, userID string) (string, error) {
	_, token, err := c.jwtauth.Encode(jwt.StandardClaims{
		ExpiresAt: time.Now().Add(expiration).Unix(),
		Id:        requestID,
		Subject:   userID,
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

func getUserID(ctx context.Context) (string, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return "", err
	}

	sub, ok := claims["sub"]
	if !ok {
		return "", errors.New("subject not found in authorization token")
	}

	userID, ok := sub.(string)
	if !ok {
		return "", errors.New("subject from authorization token is not of string type")
	}

	return userID, nil
}

// AuthOption ...
type AuthOption func(*Auth)

// AuthTokenExpiration ...
func AuthTokenExpiration(e time.Duration) AuthOption {
	return func(c *Auth) {
		c.authTokenExpiration = e
	}
}

// RefreshTokenExpiration ...
func RefreshTokenExpiration(e time.Duration) AuthOption {
	return func(c *Auth) {
		c.refreshTokenExpiration = e
	}
}

type login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type token struct {
	AuthToken    string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}
