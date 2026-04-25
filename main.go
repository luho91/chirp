package main

import (
	_ "github.com/lib/pq"
	godotenv "github.com/joho/godotenv"
	database "github.com/luho91/chirp/internal/database"
	auth "github.com/luho91/chirp/internal/auth"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries *database.Queries
	platform string
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
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		return
	}
	cfg.fileserverHits.Store(0)
	cfg.dbQueries.DeleteAllUsers(r.Context())
	/** if err != nil {
		fmt.Println("truncate fail", err)
	} **/
}

func healthz(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h["Content-Type"] = []string {"text/plain; charset=utf-8"}
	w.WriteHeader(200)
	_, _ = w.Write([]byte("OK"))
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type parameters struct {
		Password	string
		Email		string
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("decode fail", err)
		return
	}

	user, err := cfg.dbQueries.GetUser(r.Context(), params.Email)

	if err != nil {
		w.WriteHeader(404)
		fmt.Println("user not found", err)
		return
	}

	authenticated, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("check password fail", err)
		return
	}

	if !authenticated {
		w.WriteHeader(401)
		fmt.Println("forbidden")
		return
	}
	
	res, err := userJsonFromQueryData(user)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("json marshal fail", err)
		return
	}

	w.WriteHeader(200)
	w.Write(res)
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
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /api/healthz", healthz)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.metricsRead)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.metricsReset)
	serveMux.HandleFunc("POST /api/users", apiCfg.createUser)
	serveMux.HandleFunc("POST /api/chirps", apiCfg.createChirp)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirp)
	serveMux.HandleFunc("POST /api/login", apiCfg.login)

	err = server.ListenAndServe()

	if err != nil {
		fmt.Println(err)
	}
}
