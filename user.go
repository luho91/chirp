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
	ID			uuid.UUID	`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
	Email		string		`json:"email"`
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
		} else {
			w.WriteHeader(201)
			userResponse, err := userJsonFromQueryData(user)
			if err != nil {
				w.WriteHeader(500)
				fmt.Println("unmarshal fail")
				return
			}
			w.Write(userResponse)
		}
	}
}

func userJsonFromQueryData(u database.User) ([]byte, error) {
	user, err := json.Marshal(User {
		ID:			u.ID,
		CreatedAt:	u.CreatedAt,
		UpdatedAt:	u.UpdatedAt,
		Email:		u.Email,
	})

	return user, err
}


