package main

import (
	"net/http"

	"github.com/ekachaikeaw/social/internal/store"
)

// getUserFeedHandler godoc
//
//	@Summary		Fetches the user feed
//	@Description	Fetches the user feed
//	@Tags			feed
//	@Accept			json
//	@Produce		json
//	@Param			since	query		string	false	"Since"
//	@Param			until	query		string	false	"Until"
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			tags	query		string	false	"Tags"
//	@Param			search	query		string	false	"Search"
//	@Success		200		{object}	[]store.PostWithMetadata
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
func (a *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	// pagination, filter
	// create default paginatedQuery 
	fq := store.PaginatedQuery{
		Limit: 20,
		Offset: 0,
		Sort: "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return 
	}

	if err := Validate.Struct(fq); err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	c := r.Context()
	user := getUserFromCtx(r)

	feed, err := a.store.Posts.GetUserFeed(c, user.ID, fq)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err = a.jsonResponse(w, http.StatusOK, feed); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}