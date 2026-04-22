package main

import(
	"time"
	"github.com/google/uuid"
)

type Chirp struct {
	ID			uuid.UUID	`json:"id"`
	UserID		uuid.UUID	`json:"user_id"`
	Body		string		`json:"body"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
}
