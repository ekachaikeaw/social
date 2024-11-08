package config

import "net/http"

func (a *Application) HealthCheckHandler(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("OK"))
}