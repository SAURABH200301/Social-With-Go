package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/SAURABH200301/Social/internal/store"
	"github.com/go-chi/chi/v5"
)

type CreatePostPayload struct {
	Content string   `json:"content" validate:"required"`
	Title   string   `json:"title" validate:"required"`
	Tags    []string `json:"tags,omitempty"`
}

// Create Post Handler
//	@Summary		Create a new post
//	@Description	Creates a new post with the provided content, title, and optional tags.
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			post	body		CreatePostPayload	true	"Post Payload"
//	@Success		201		{object}	store.Post
//	@Failure		400		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Security		BearerAuth
//	@Router			/posts [post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var postPayload CreatePostPayload
	if err := readJSON(w, r, &postPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(postPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// uid, ok := r.Context().Value("userID").(int64)
	// if !ok {
	// 	_ = writeErrorJSON(w, http.StatusUnauthorized, "missing user id")
	// 	return
	// }

	post := store.Post{
		UserID:    1,
		Content:   postPayload.Content,
		Title:     postPayload.Title,
		CreatedAt: time.Now().Format(time.RFC3339),
		Tags:      postPayload.Tags,
	}
	err := app.store.Posts.Create(r.Context(), &post)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	err = writeJSON(w, http.StatusCreated, post)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// Get Post Handler
//	@Summary		Get a post by ID
//	@Description	Retrieves a post by its ID, including associated comments.
//	@Tags			Posts
//	@Produce		json
//	@Param			postID	path		int	true	"Post ID"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	errorResponse
//	@Failure		404		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Security		BearerAuth
//	@Router			/posts/{postID} [get]
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	comments, err := app.store.Comments.GetCommentsByPostID(r.Context(), int32(post.ID))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = make([]store.Comments, len(comments))
	for i, comment := range comments {
		post.Comments[i] = *comment
	}

	err = app.jsonResponse(w, http.StatusOK, post)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// Delete Post Handler
//	@Summary		Delete a post by ID
//	@Description	Deletes a post by its ID.
//	@Tags			Posts
//	@Param			postID	path	int	true	"Post ID"
//	@Success		204
//	@Failure		400	{object}	errorResponse
//	@Failure		404	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Security		BearerAuth
//	@Router			/posts/{postID} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	postIDParam := chi.URLParam(r, "postID")
	postID, err := strconv.ParseInt(postIDParam, 10, 32)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	err = app.store.Posts.DeletePostByID(r.Context(), int32(postID))
	if err != nil {
		app.notFoundResponse(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type UpdatePayload struct {
	Title   *string   `json:"title,omitempty"`
	Content *string   `json:"content,omitempty"`
	Tags    *[]string `json:"tags,omitempty"`
}

// Update Post Handler
//	@Summary		Update a post by ID
//	@Description	Updates a post's title, content, and/or tags by its ID.
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int				true	"Post ID"
//	@Param			post	body		UpdatePayload	true	"Update Payload"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	errorResponse
//	@Failure		404		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Security		BearerAuth
//	@Router			/posts/{postID} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Tags != nil {
		post.Tags = *payload.Tags
	}

	if err := app.store.Posts.UpdatePost(r.Context(), post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
func (app *application) jsonResponse(w http.ResponseWriter, status int, data interface{}) error {
	type jsonResponse struct {
		Data interface{} `json:"data"`
	}
	return writeJSON(w, status, jsonResponse{Data: data})
}

// MIDDLEWARE TO FETCH POST AND ADD TO CONTEXT
type postKey string

const POST_CTX_KEY postKey = "post"

func (app *application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postIDParam := chi.URLParam(r, "postID")
		postID, err := strconv.ParseInt(postIDParam, 10, 32)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
		ctx := r.Context()
		post, err := app.store.Posts.GetByID(ctx, int32(postID))
		if err != nil {
			app.notFoundResponse(w, r, err)
			return

		}
		ctx = context.WithValue(ctx, POST_CTX_KEY, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value("post").(*store.Post)
	return post
}
