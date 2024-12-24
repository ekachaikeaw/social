package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ekachaikeaw/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type userKey struct{}

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (a *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	c := r.Context()

	user, err := a.getUser(c, userID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			a.notFoundErr(w, r, err)
			return
		default:
			a.internalServerError(w, r, err)
			return
		}
	}
	a.jsonResponse(w, http.StatusOK, user)
}

type UserPayload struct {
	UserID int64 `json:"user_id"`
}

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User followed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
func (a *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromCtx(r)

	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	c := r.Context()
	err = a.store.Follower.Follow(c, followerUser.ID, followedID)
	if err != nil {
		switch err {
		case store.ErrConflict:
			a.conflictErr(w, r, err)
			return
		default:
			a.internalServerError(w, r, err)
			return
		}
	}

	if err := a.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

// UnfollowUser gdoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User unfollowed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/unfollow [put]
func (a *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	follower := getUserFromCtx(r)

	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	c := r.Context()

	err = a.store.Follower.Unfollow(c, follower.ID, unfollowedID)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ActivateUser godoc
//
//	@Summary		Activates/Register a user
//	@Description	Activates/Register a user by invitation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (a *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	if err := a.store.Users.Activate(r.Context(), token); err != nil {
		switch err {
		case store.ErrNotFound:
			a.notFoundErr(w, r, err)
		default:
			a.internalServerError(w, r, err)
		}
		return
	}

	// if err := a.jsonResponse(w, http.StatusNoContent, ""); err != nil {
	// 	a.internalServerError(w, r, err)
	// }
	w.WriteHeader(http.StatusNoContent)
}

func (a *application) getUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			a.internalServerError(w, r, err)
			return
		}

		c := r.Context()

		user, err := a.store.Users.GetByID(c, userID)
		if err != nil {
			switch err {
			case store.ErrNotFound:
				a.notFoundErr(w, r, err)
				return
			default:
				a.internalServerError(w, r, err)
				return
			}
		}

		c = context.WithValue(c, userKey{}, user)
		next.ServeHTTP(w, r.WithContext(c))
	})
}

func getUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userKey{}).(*store.User)
	return user
}
