package main

import (
	"encoding/json"
	"net/http"
)

func handleChirpValidation(w http.ResponseWriter, r *http.Request) {
	
	
	
	type rqst struct {
		Body string `json:"body"`
	}
	type success struct {
		Valid bool `json:"valid"`
	}

	var rqstBody rqst
	decode := json.NewDecoder(r.Body)
	if err := decode.Decode(&rqstBody); err != nil {
		respondWithError(w, 400, "Issue in decoding request body")
	}

	if len(rqstBody.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
	}

	respondWithJSON(w, http.StatusOK, success{
		Valid: true,
	})


}