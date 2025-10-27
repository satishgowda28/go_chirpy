package main

import (
	"fmt"
	"net/http"
)


func (cfg *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {

	chirps, err := cfg.db.ListChirps(r.Context())
	if err != nil {
		errStr := fmt.Sprintf("Issue on fetching chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, errStr)
	}

	allChirps := []Chirp{}
	for _, dbChirp := range chirps {
		allChirps = append(allChirps, Chirp{ID: dbChirp.ID,Body: dbChirp.Body,CreatedAt: dbChirp.CreatedAt, UpdatedAt: dbChirp.CreatedAt, UserID: dbChirp.UserID})
	}

	respondWithJSON(w, http.StatusOK, allChirps)

}