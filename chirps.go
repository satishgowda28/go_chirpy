package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/satishgowda28/go_chirpy/internal/auth"
	"github.com/satishgowda28/go_chirpy/internal/database"
)

type Chirp struct {
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body string `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Body string `json:"body"`
		// UserID uuid.UUID `json:"user_id"`
	}

	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(authToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}


	params := parameter{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Issue in decoding the params")
		return
	}

	cleanBody, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{Body: cleanBody, UserID: userID})
	if err != nil {
		errstr := fmt.Sprintf("issue in inserting in table: %s",err)
		respondWithError(w, http.StatusInternalServerError, errstr)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})

}

func validateChirp(body string) (string, error) {
	if len(body) > 140 {
		return "", errors.New("chirp is too long")
	}
	profaneList := map[string]struct{}{
		"kerfuffle": {},
		"sharbert": {},
		"fornax":{},
	}
	cleaned := getCleanedBody(body, profaneList)
	return cleaned, nil
}

func getCleanedBody(body string, profane map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		lowerdWord := strings.ToLower(word)
		if _, ok := profane[lowerdWord]; ok {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

