package main

import (
	_ "github.com/lib/pq"
	database "github.com/luho91/chirp/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries *database.Queries
}

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
	cfg.fileserverHits.Store(0)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h["Content-Type"] = []string {"text/plain; charset=utf-8"}
	w.WriteHeader(200)
	_, _ = w.Write([]byte("OK"))
}

func (cfg *apiConfig) validateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		Error string `json:"error"`
		Valid bool `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}

	naziGermany := []string {"kerfuffle", "sharbert", "fornax"}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	respBody := returnVals{}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(500)
		respBody.Error = "Something went wrong during JSON decode"
	} else if len(params.Body) > 140 {
		w.WriteHeader(400)
		respBody.Error = "Chirp is too long"
	} else {
		respBody.Valid = true
		w.WriteHeader(200)
		words := strings.Split(params.Body, " ")
		for _, toCensor := range naziGermany {
			for i, word := range words {
				if strings.ToLower(word) == toCensor {
					words[i] = "****"
				}
			}
		}
		respBody.CleanedBody = strings.Join(words, " ")
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		respBody.Error = "Something went wrong during JSON encode"
		w.WriteHeader(500)
		return
	}

	w.Write(dat)

}

func main() {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		fmt.Println(err)
	}

	dbQueries := database.New(db)
	serveMux := http.NewServeMux()
	server := http.Server{}
	server.Handler = serveMux
	server.Addr = ":8080"
	apiCfg := apiConfig{}
	apiCfg.dbQueries = dbQueries
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /api/healthz", healthz)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.metricsRead)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.metricsReset)
	serveMux.HandleFunc("POST /api/validate_chirp", apiCfg.validateChirp)

	err = server.ListenAndServe()

	if err != nil {
		fmt.Println(err)
	}
}
