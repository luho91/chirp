package main

import(
	_ "github.com/lib/pq"
	database "github.com/luho91/chirp/internal/database"
	auth "github.com/luho91/chirp/internal/auth"
	"github.com/google/uuid"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"strings"
)

type Chirp struct {
	ID			uuid.UUID	`json:"id"`
	UserID		uuid.UUID	`json:"user_id"`
	Body		string		`json:"body"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
}

func validateChirp(chirpContent string) (naziGermanyConformMessage string, isValidLength bool) {
	naziGermany := []string {"kerfuffle", "sharbert", "fornax"}

	if len(chirpContent) > 140 {
		return chirpContent, false
	} else {
		words := strings.Split(chirpContent, " ")
		for _, toCensor := range naziGermany {
			for i, word := range words {
				if strings.ToLower(word) == toCensor {
					words[i] = "****"
				}
			}
		}
		return strings.Join(words, " "), true
	}
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body		string		`json:"body"`
		Token		string		`json:"token"`
		UserID		uuid.UUID	`json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("decode fail", err)
		return
	}

	newBody, isValidLength := validateChirp(params.Body)
	if !isValidLength {
		w.WriteHeader(500)
		return
	}
	params.Body = newBody

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("token fail", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.appSecret)

	if err != nil {
		w.WriteHeader(401)
		fmt.Println("token validation fail", err, cfg.appSecret)
		return
	}

	params.UserID = userID
	
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams {
		UserID:	params.UserID,
		Body:	params.Body,
	})
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("create chirp fail", err)
		return
	} else {
		w.WriteHeader(201)
		respBody := Chirp {
			ID:			chirp.ID,
			CreatedAt: 	chirp.CreatedAt,
			UpdatedAt:	chirp.UpdatedAt,
			Body:		chirp.Body,
			UserID:		chirp.UserID,
		}
		data, err := json.Marshal(respBody)
		if err != nil {
			w.WriteHeader(500)
			fmt.Println("unmarshal fail", err)
			return
		}
		w.Write(data)
	}
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	authorFilter := r.URL.Query().Get("author_id")

	aF := uuid.UUID{}

	if authorFilter != "" {
		aF, _ = uuid.Parse(authorFilter)
	}

	chirps, err := cfg.dbQueries.GetChirps(r.Context(), aF)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("get chirp fail", err)
		return
	}


	w.WriteHeader(200)
	outChirps := []Chirp {}
	for _, chirp := range chirps {
		outChirps = append(outChirps, Chirp {
			ID:			chirp.ID,
			UserID:		chirp.UserID,
			Body:		chirp.Body,
			CreatedAt:	chirp.CreatedAt,
			UpdatedAt:	chirp.UpdatedAt,
		})
	}

	data, err := json.Marshal(outChirps)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("unmarshal fail", err)
		return
	}
	w.Write(data)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	chirpId, err := uuid.Parse(r.PathValue("chirpID"))

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("string to uuid fail", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpId)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(404)
			fmt.Println("chirp not found")
			return
		}
		w.WriteHeader(500)
		fmt.Println("get chirp fail", err)
		return
	}

	if chirp.ID == uuid.Nil {
	}

	w.WriteHeader(200)
	outChirp := Chirp {
		ID:			chirp.ID,
		UserID:		chirp.UserID,
		Body:		chirp.Body,
		CreatedAt:	chirp.CreatedAt,
		UpdatedAt:	chirp.UpdatedAt,
	}

	data, err := json.Marshal(outChirp)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("unmarshal fail", err)
		return
	}
	w.Write(data)
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpId, err := uuid.Parse(r.PathValue("chirpID"))

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpId)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("get chirp fail", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	
	if err != nil {
		w.WriteHeader(401)
		fmt.Println("token fail", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.appSecret)

	if userId != chirp.UserID {
		w.WriteHeader(403)
		fmt.Println("unauthorized")
		return
	}

	if err != nil {
		w.WriteHeader(401)
		fmt.Println("token validation fail", err)
		return
	}

	err = cfg.dbQueries.DeleteChirp(r.Context(), chirpId)

	if err != nil {
		w.WriteHeader(404)
		fmt.Println("not found?", err)
		return
	}

	w.WriteHeader(204)
}

