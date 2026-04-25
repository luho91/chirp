package main

import (
	_ "github.com/lib/pq"
	godotenv "github.com/joho/godotenv"
	database "github.com/luho91/chirp/internal/database"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"database/sql"
)

type apiConfig struct {
	fileserverHits	atomic.Int32
	dbQueries		*database.Queries
	platform		string
	appSecret		string
}

func main() {
	godotenv.Load()
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
	apiCfg.platform = os.Getenv("PLATFORM")
	apiCfg.appSecret = os.Getenv("APP_SECRET")
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /api/healthz", healthz)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.metricsRead)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.metricsReset)
	serveMux.HandleFunc("POST /api/users", apiCfg.createUser)
	serveMux.HandleFunc("POST /api/chirps", apiCfg.createChirp)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirp)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirp)
	serveMux.HandleFunc("POST /api/login", apiCfg.login)
	serveMux.HandleFunc("POST /api/refresh", apiCfg.refresh)
	serveMux.HandleFunc("POST /api/revoke", apiCfg.revoke)
	serveMux.HandleFunc("PUT /api/users", apiCfg.putUser)
	serveMux.HandleFunc("POST /api/polka/webhooks", apiCfg.premiumUser)

	err = server.ListenAndServe()

	if err != nil {
		fmt.Println(err)
	}
}
