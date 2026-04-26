package main

import(
	database "github.com/luho91/chirp/internal/database"
	auth "github.com/luho91/chirp/internal/auth"
	"time"
	"github.com/google/uuid"
	"net/http"
	"encoding/json"
	"fmt"
)

type User struct {
	ID				uuid.UUID	`json:"id"`
	CreatedAt		time.Time	`json:"created_at"`
	UpdatedAt		time.Time	`json:"updated_at"`
	Email			string		`json:"email"`
	Token			string		`json:"token"`
	RefreshToken	string		`json:"refresh_token"`
	ChirpyRed		bool		`json:"is_chirpy_red"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}

	err := decoder.Decode(&params)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("decode fail", err)
		return
	} else {
		hashedPassword, err := auth.HashPassword(params.Password)

		if err != nil {
			w.WriteHeader(500)
			fmt.Println("hash fail", err)
			return
		}

		user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams {
			Email:			params.Email,
			HashedPassword:	hashedPassword,
		})
		if err != nil {
			w.WriteHeader(500)
			fmt.Println("create user fail", err)
			return
		}

		refreshTokenString := auth.MakeRefreshToken()

		rParams := database.CreateRefreshTokenParams {}

		rParams.UserID = user.ID
		rParams.Token = refreshTokenString
		rParams.ExpiresAt = time.Now().UTC().Add(time.Duration(60 * 24 * time.Hour))

		rToken, err := cfg.dbQueries.CreateRefreshToken(r.Context(), rParams)

		if err != nil {
			w.WriteHeader(500)
			fmt.Println("refresh token fail", err)
			return
		} else {
			w.WriteHeader(201)
			userResponse, err := userJsonFromQueryData(user, rToken)
			if err != nil {
				w.WriteHeader(500)
				fmt.Println("unmarshal fail")
				return
			}
			w.Write(userResponse)
		}
	}
}

func (cfg *apiConfig) putUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		UserID		uuid.UUID	`json:"user_id"`
		Password	string		`json:"password"`
		Email		string		`json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("decode fail", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		w.WriteHeader(401)
		fmt.Println("token fail", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.appSecret)

	if err != nil {
		w.WriteHeader(401)
		fmt.Println("token validation fail", err)
		return
	}

	params.UserID = userID
	hashedPassword, err := auth.HashPassword(params.Password)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("password hash fail", err)
		return
	}

	res := database.UpdateUserParams {
		ID:				userID,
		HashedPassword:	hashedPassword,
		Email:			params.Email,
	}

	user, err := cfg.dbQueries.UpdateUser(r.Context(), res)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("update user query error", err)
		return
	}
	refreshTokenString := auth.MakeRefreshToken()

	rParams := database.CreateRefreshTokenParams {}

	rParams.UserID = user.ID
	rParams.Token = refreshTokenString
	rParams.ExpiresAt = time.Now().UTC().Add(time.Duration(60 * 24 * time.Hour))

	rToken, err := cfg.dbQueries.CreateRefreshToken(r.Context(), rParams)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("refresh token fail", err)
		return
	}

	w.WriteHeader(200)
	userResponse, err := userJsonFromQueryData(user, rToken)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("unmarshal fail")
		return
	}
	w.Write(userResponse)	
}

func (cfg *apiConfig) premiumUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event		string		`json:"event"`
		Data		struct {
			UserID	string		`json:"user_id"`
		} 						`json:"data"`
	}

	key, err := auth.GetAPIKey(r.Header)

	if err != nil || key != cfg.polkaKey {
		w.WriteHeader(401)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("decode fail", err)
		return
	}
	
	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	u, err := uuid.Parse(params.Data.UserID)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println("uuid from string fail", err)
		return
	}

	err = cfg.dbQueries.PremiumUser(r.Context(), u)

	if err != nil {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(204)
}

func userJsonFromQueryData(u database.User, r database.RefreshToken) ([]byte, error) {
	user, err := json.Marshal(User {
		ID:				u.ID,
		CreatedAt:		u.CreatedAt,
		UpdatedAt:		u.UpdatedAt,
		Email:			u.Email,
		Token:			u.JwtToken.String,
		RefreshToken:	r.Token,
		ChirpyRed:		u.IsChirpyRed,
	})

	return user, err
}


