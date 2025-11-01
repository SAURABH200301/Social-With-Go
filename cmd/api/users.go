package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// GetUser godoc
//
// @Summary		Get User by ID
// @Description	Get a user by their unique ID.
// @Tags			Users
// @Accept			json
// @Produce		json
// @Param			userID	path		int	true	"User ID"
// @Success		200		{object}	store.Users
// @Failure		400		{object}	errorResponse
// @Failure		500		{object}	errorResponse
// @Router			/users/{userID} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 32)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	user, err := app.store.Users.GetByID(ctx, int64(userID))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	err = app.jsonResponse(w, http.StatusOK, user)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
