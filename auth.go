package main

import (
	"encoding/json"
	"fmt"
	"time"
	"net/http"
	database "github.com/luho91/chirp/internal/database"
	auth "github.com/luho91/chirp/internal/auth"
	"database/sql"
	"strings"
)

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type parameters struct {
		Password			string	`json:"password"`
		Email				string	`json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("decode fail", err)
		return
	}

	expireTime := time.Duration(60 * 60) * time.Second

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

	token, err := auth.MakeJWT(user.ID, cfg.appSecret, expireTime)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("make token fail", err)
		return
	}

	user.JwtToken = sql.NullString {
		String:	token,
		Valid:	true,
	}

	refreshTokenString := auth.MakeRefreshToken()
	
	rParams := database.CreateRefreshTokenParams {}
	
	rParams.UserID = user.ID
	rParams.Token = refreshTokenString
	rParams.ExpiresAt = time.Now().UTC().Add(time.Duration(60 * 24 * time.Hour))

	rToken, err := cfg.dbQueries.CreateRefreshToken(r.Context(), rParams)

	res, err := userJsonFromQueryData(user, rToken)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("json marshal fail", err)
		return
	}

	w.WriteHeader(200)
	w.Write(res)
}

func (cfg *apiConfig) refresh(w http.ResponseWriter, r *http.Request) {

	rTokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	token, err := cfg.dbQueries.GetRefreshToken(r.Context(), rTokenString)

	if err != nil || token.RevokedAt.Valid || token.ExpiresAt.Before(time.Now().UTC()) {
		w.WriteHeader(401)
		fmt.Printf("rToken: %s\n", rTokenString)
		fmt.Println("token revoked or expired or error", err)
		return
	}

	expireTime := time.Duration(60 * 60) * time.Second
	JWT, err := auth.MakeJWT(token.UserID, cfg.appSecret, expireTime)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("make token fail", err)
		return
	}

	res, err := json.Marshal(map[string]string{
		"token": JWT,
	})
	
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("token marshal fail", err)
		return
	}

	w.WriteHeader(200)
	w.Write(res)
}

func (cfg *apiConfig) revoke(w http.ResponseWriter, r *http.Request) {
	rTokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	token, err := cfg.dbQueries.GetRefreshToken(r.Context(), rTokenString)

	if err != nil || token.RevokedAt.Valid || token.ExpiresAt.Before(time.Now().UTC()) {
		w.WriteHeader(401)
		fmt.Printf("rToken: %s\n", rTokenString)
		fmt.Println("token revoked or expired or error", err)
		return
	}

	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), token.ID)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("token revoke error", err)
		return
	}

	w.WriteHeader(204)
}

