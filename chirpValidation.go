package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handleChirpValidation(w http.ResponseWriter, r *http.Request) {

	/** this is what i came up with string slice */
	// profaneList := []string{"kerfuffle","sharbert", "fornax"}

	/* This is solution from the solution of the site to use map */
	profaneList := map[string]struct{}{
		"kerfuffle": {},
		"sharbert": {},
		"fornax":{},
	}
	
	type rqst struct {
		Body string `json:"body"`
	}
	type success struct {
		CleanedBody string `json:"cleaned_body"`
	}

	var rqstBody rqst
	decode := json.NewDecoder(r.Body)
	if err := decode.Decode(&rqstBody); err != nil {
		respondWithError(w, 400, "Issue in decoding request body")
		return
	}

	if len(rqstBody.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	words := strings.Split(rqstBody.Body, " ")
	for i, word := range words {
			lowerdWord := strings.ToLower(word)
			if _, ok := profaneList[lowerdWord]; ok {
				words[i] = "****"
			}
		}

	respondWithJSON(w, http.StatusOK, success{
		CleanedBody: strings.Join(words, " "),
	})


}