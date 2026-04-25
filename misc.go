package main

import (
	"net/http"
	"fmt"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsRead(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	h := w.Header()
	h["Content-Type"] = []string {"text/html"}
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) metricsReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		return
	}
	cfg.fileserverHits.Store(0)
	cfg.dbQueries.DeleteAllUsers(r.Context())
}

func healthz(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h["Content-Type"] = []string {"text/plain; charset=utf-8"}
	w.WriteHeader(200)
	_, _ = w.Write([]byte("OK"))
}
