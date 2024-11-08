package config

import (
	"log"
	"net/http"
	"time"

	"github.com/ekachaikeaw/social/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Application struct {
	Config Config
	Store  store.Storage
}

type DbConfig struct {
	Addr        string
	MaxOpenConn int
	MaxIdleConn int
	MaxIdleTime string
}

type Config struct {
	Addr string
	Db   DbConfig
}

func (a *Application) Mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", a.HealthCheckHandler)
	})

	return r
}

func (a *Application) Run(mux http.Handler) error {
	server := &http.Server{
		Addr:         a.Config.Addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	log.Printf("server listenning at %s", a.Config.Addr)

	return server.ListenAndServe()
}
