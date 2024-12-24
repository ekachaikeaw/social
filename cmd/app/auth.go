package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/ekachaikeaw/social/internal/mailer"
	"github.com/ekachaikeaw/social/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// registerUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/user [post]
func (a *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJson(w, r, &payload); err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		Role: store.Role{
			Name: "user",
		},
	}

	// hash password
	if err := user.Password.Set(payload.Password); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	c := r.Context()

	// hash the token to storage but plain token for email
	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	err := a.store.Users.CreateAndInvite(c, user, hashToken, a.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			a.badRequestResponse(w, r, err)
		case store.ErrDuplicateUsername:
			a.badRequestResponse(w, r, err)
		default:
			a.internalServerError(w, r, err)
		}
		return
	}

	// create email
	isProdEnv := a.config.env == "production"
	activationURL := fmt.Sprintf("%s/confirm/%s", a.config.frontedURL, plainToken)

	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}
	status, err := a.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		// if fail rollback
		a.logger.Errorw("error sending welcome email", "error", err)

		// rollback user creation if email fails (SAGA pattern)
		if err := a.store.Users.Delete(c, user.ID); err != nil {
			a.logger.Errorw("error deleting user", "error", err)
		}

		a.internalServerError(w, r, err)
		return
	}

	a.logger.Infow("Email sent", "status code", status, "error", err)

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}
	if err := a.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

// createTokenHandler godoc
//
//	@Summary		Creates a token
//	@Description	Creates a token for a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserTokenPayload	true	"User credentials"
//	@Success		200		{string}	string					"Token"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/token [post]
func (a *application) createUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	// parse payload credentials
	var payload CreateUserTokenPayload
	if err := readJson(w, r, &payload); err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// fetch user from email
	user, err := a.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			a.unauthorizedErrorResponse(w, r, err)
		default:
			a.internalServerError(w, r, err)
		}
		return
	}

	if err := user.Password.Compare(payload.Password); err != nil {
		a.unauthorizedErrorResponse(w, r, err)
		return
	}
	// create token
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(a.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": a.config.auth.token.iss,
		"aud": a.config.auth.token.iss,
	}
	token, err := a.authenticator.GenerateToken(claims)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}
	// send to user
	if err := a.jsonResponse(w, http.StatusCreated, token); err != nil {
		a.internalServerError(w, r, err)
	}
}
