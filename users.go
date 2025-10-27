package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpadatedAt time.Time `json:"updated_at"`
	Email string `json:"email"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type newUserData struct {
		Email string `json:"email"`
	}
	type response struct {
		User
	}

	var newUser newUserData
	decode := json.NewDecoder(r.Body)
	if err := decode.Decode(&newUser); err != nil {
		respondWithError(w, 400, "Issue in decode the request body")
		return
	}

	if newUser.Email == "" {
		respondWithError(w, 400, "Email cannot be empty")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), newUser.Email)
	if err != nil {
		errorStr := fmt.Sprintf("Issues in creating users: %s", err)
		respondWithError(w, 400, errorStr)
		return
	}

	respondWithJSON(w, 201, response{User: User{ID: user.ID, CreatedAt: user.CreatedAt, UpadatedAt: user.UpdatedAt, Email: user.Email}})
}