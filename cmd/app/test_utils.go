package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ekachaikeaw/social/internal/auth"
	"github.com/ekachaikeaw/social/internal/store"
	"github.com/ekachaikeaw/social/internal/store/cache"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T) *application {
	t.Helper()

	logger := zap.NewNop().Sugar()
	// logger := zap.Must(zap.NewProduction()).Sugar()
	mockStore := store.NewMockStore()
	mockCache := cache.NewMockStore()
	mockAuth := auth.NewMockAuth()
	return &application{
		logger:        logger,
		store:         mockStore,
		cacheStore:    mockCache,
		authenticator: mockAuth,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("expected %d. Got %d", expected, actual)
	}
}
