package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// GetUser godoc
//
//	@Summary		Get User by ID
//	@Description	Get a user by their unique ID.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int	true	"User ID"
//	@Success		200		{object}	store.Users
//	@Failure		400		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/users/{userID} [get]
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

// Activate User Handler
//	@Summary		Activate a user account
//	@Description	Activates a user account using the provided activation token.
//	@Tags			Users
//	@Param			token	path	string	true	"Activation Token"
//	@Success		204
//	@Failure		400	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusNoContent, ""); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
