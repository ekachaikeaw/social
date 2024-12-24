package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/ekachaikeaw/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type postKey struct{}
type PostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=100"`
}

// CreatePost godoc
//
//	@Summary		Creates a post
//	@Description	Creates a post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		PostPayload	true	"Post payload"
//	@Success		201		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (a *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload PostPayload

	if err := readJson(w, r, &payload); err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	user := getUserFromCtx(r)
	post := &store.Post{
		Content: payload.Content,
		Title:   payload.Title,
		Tags:    payload.Tags,
		UserID:  user.ID,
	}

	if err := a.store.Posts.Create(ctx, post); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := a.jsonResponse(w, http.StatusCreated, post); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

// GetPost godoc
//
//	@Summary		Fetches a post
//	@Description	Fetches a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	store.Post
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [get]
func (a *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	comments, err := a.store.Comment.GetByPostID(r.Context(), post.ID)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := a.jsonResponse(w, http.StatusOK, post); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

// DeletePost godoc
//
//	@Summary		Deletes a post
//	@Description	Delete a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		204	{object} string
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [delete]
func (a *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "postID")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	c := r.Context()
	err = a.store.Posts.Delete(c, idInt)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			a.notFoundErr(w, r, err)
		default:
			a.internalServerError(w, r, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdatePost godoc
//
//	@Summary		Updates a post
//	@Description	Updates a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostPayload	true	"Post payload"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [patch]
func (a *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePostPayload
	if err := readJson(w, r, &payload); err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}

	c := r.Context()
	err := a.updatePost(c, post)
	
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			a.conflictErr(w, r, err)	
			return
		default:
			a.internalServerError(w, r, err)
			return
		}
	}

	if err := a.jsonResponse(w, http.StatusOK, post); err != nil {
		a.internalServerError(w, r, err)
	}
}

func (a *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "postID")
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			a.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()
		post, err := a.store.Posts.GetByID(ctx, idInt)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrConflict):
				a.conflictErr(w, r, err)
				return
			default:
				a.internalServerError(w, r, err)
				return
			}
		}

		ctx = context.WithValue(ctx, postKey{}, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postKey{}).(*store.Post)
	return post
}

func (a *application)updatePost(c context.Context, post *store.Post) error {
	if err := a.store.Posts.Update(c , post); err != nil {
		return err
	}

	a.cacheStore.Users.Delete(c, post.UserID)
	return nil
}