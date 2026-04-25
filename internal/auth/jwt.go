package auth

import (
	"github.com/google/uuid"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"errors"
	"strings"
	"fmt"
)

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims {
		Issuer:		"chirpy-access",
		IssuedAt:	jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt:	jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:	userId.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tokenSecret))

	return tokenString, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		fmt.Println("parse with claims fail", err)
		return uuid.Nil, err
	}

	userID, err := token.Claims.GetSubject()

	if err != nil {
		fmt.Println("get subject fail", err)
		return uuid.Nil, err
	}

	parsedUserID, err := uuid.Parse(userID)

	return parsedUserID, err
}

func GetBearerToken(headers http.Header) (string, error) {
	h := headers.Get("Authorization")

	if h == "" {
		return "", errors.New("No bearer token provided")
	}

	tokenString := strings.TrimPrefix(h, "Bearer ")		
	
	return tokenString, nil	
}
