package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/satishgowda28/go_chirpy/internal/auth"
	"github.com/satishgowda28/go_chirpy/internal/database"
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
		Password string `json:"password"`
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

	if newUser.Password == "" {
		respondWithError(w, 400, "Email cannot be empty")
		return
	}

	hashedPassword, err := auth.HashPassord(newUser.Password)
	if err != nil {
		respondWithError(w, 400, "Password cannot be saved")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email: newUser.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		errorStr := fmt.Sprintf("Issues in creating users: %s", err)
		respondWithError(w, 400, errorStr)
		return
	}

	respondWithJSON(w, 201, response{User: User{ID: user.ID, CreatedAt: user.CreatedAt, UpadatedAt: user.UpdatedAt, Email: user.Email}})
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email string `json:"Email"`
		Password string `json:"password"`
	}

	

	loginParams := params{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&loginParams); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user, err := cfg.db.GetUserByEmail(r.Context(), loginParams.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Issue with user email id")
		return
	}

	mached, _ := auth.CheckPasswordHash(loginParams.Password, user.HashedPassword)
	if !mached  {
		respondWithError(w, http.StatusUnauthorized, "Use is not authorized")
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID: user.ID, CreatedAt: user.CreatedAt, UpadatedAt: user.UpdatedAt, Email: user.Email,
	})


}