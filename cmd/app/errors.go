package main

import (
	"net/http"
)

func (a *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusInternalServerError, "the server encounter a problem")
}

func (a *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	a.logger.Warnw("forbidden", "method", r.Method, "path", r.URL.Path, "error")

	writeJSONError(w, http.StatusForbidden, "forbidden")
}

func (a *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.Warnw("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (a *application) notFoundErr(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.Errorw("not found", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusNotFound, "Not Found")
}

func (a *application) conflictErr(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.Errorw("conflict", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusConflict, err.Error())
}

func (a *application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (a *application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.Warnf("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}