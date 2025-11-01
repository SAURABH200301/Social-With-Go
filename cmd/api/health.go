package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Status  string `json:"status"`
		Env     string `json:"env,omitempty"`
		Version string `json:"version,omitempty"`
	}{
		Status:  "ok ",
		Env:     app.config.env,
		Version: version,
	}
	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
