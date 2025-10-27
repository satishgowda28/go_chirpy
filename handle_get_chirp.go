package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("chirpID")
	id, err := uuid.Parse(chirpId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "id format is incorrect")
		return 
	}
	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, fmt.Sprintf("chirp with id %s not found", id))	
			return 
		}
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Issue in retreving chirp of id :%s", id))
		return 
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.CreatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	})
}