package main

import (
	"net/http"

	"github.com/SAURABH200301/Social/internal/store"
)

//	@Summary		Get User Feed
//	@Description	Retrieve the feed for the authenticated user with pagination support.
//	@Tags			Feeds
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int		false	"Number of posts to return"				default(1)		minimum(1)	maximum(100)
//	@Param			offset	query		int		false	"Number of posts to skip"				default(0)		minimum(0)
//	@Param			sort	query		string	false	"Sort order of posts by creation time"	default(desc)	Enum(asc, desc)
//	@Success		200		{object}	store.PostWithMetadata
//	@Failure		400		{object}	errorResponse	"Bad Request"
//	@Failure		500		{object}	errorResponse	"Internal Server Error"
//	@Security		BearerAuth
//	@Router			/feeds [get]
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {

	pq := store.PaginationFeedQuery{
		Limit:  1,
		Offset: 0,
		Sort:   "desc",
	}
	fq, err := pq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	ctx := r.Context()
	userId := 1 //TO BE FETCHED FROM AUTH CONTEXT LATER
	feed, err := app.store.Posts.GetUserFeed(ctx, userId, fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	err = app.jsonResponse(w, http.StatusOK, feed)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
