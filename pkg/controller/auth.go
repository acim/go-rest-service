package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	arcmw "github.com/acim/arc/pkg/middleware"
	"github.com/acim/arc/pkg/store"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// Errors.
var (
	ErrNoSubject   = errors.New("subject not found")
	ErrInvalidType = errors.New("invalid type")
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
		authTokenExpiration:    15 * time.Minute, //nolint:gomnd
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
	res := arcmw.ResponseFromContext(r.Context())

	l := &login{} //nolint:exhaustivestruct

	err := json.NewDecoder(r.Body).Decode(l)
	if err != nil {
		c.logger.Warn("login", zap.NamedError("json decode", err))
		res.SetStatusBadRequest(errParsingRequestBody)

		return
	}

	u, err := c.users.FindByEmail(r.Context(), l.Email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			res.SetStatusForbidden(errInvalidCredentials)

			return
		}

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
	res := arcmw.ResponseFromContext(r.Context())

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

// Logout handles /auth/logout endpoint.
func (c *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	res := arcmw.ResponseFromContext(r.Context())

	userID, err := getUserID(r.Context())
	if err != nil {
		c.logger.Warn("user", zap.NamedError("get user id", err))
		res.SetStatusInternalServerError(err.Error())

		return
	}

	res.SetStatus(http.StatusNoContent)
	c.logger.Info("logout", zap.String("user id", userID))
}

func (c *Auth) token(expiration time.Duration, requestID, userID string) (string, error) {
	_, token, err := c.jwtauth.Encode(jwt.RegisteredClaims{ //nolint:exhaustivestruct
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		ID:        requestID,
		Subject:   userID,
	})
	if err != nil {
		return "", fmt.Errorf("encode token: %w", err)
	}

	return token, nil
}

func getUserID(ctx context.Context) (string, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("jwt token parse: %w", err)
	}

	sub, ok := claims["sub"]
	if !ok {
		return "", fmt.Errorf("jwt token: %w", ErrNoSubject)
	}

	userID, ok := sub.(string)
	if !ok {
		return "", fmt.Errorf("jwt token subject: %w", ErrInvalidType)
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
	RefreshToken string `json:"refreshToken,omitempty"`
}
