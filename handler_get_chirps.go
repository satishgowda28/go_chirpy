package main

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/google/uuid"
)

func (cfg *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("author_id")

	userId := uuid.Nil
	if id != "" {
		uId, err := uuid.Parse(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		}
		userId = uId
	}

	fmt.Printf("User id %s", userId)

	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		errStr := fmt.Sprintf("Issue on fetching chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, errStr)
	}

	allChirps := []Chirp{}
	for _, dbChirp := range chirps {
		if userId != uuid.Nil && dbChirp.UserID != userId {
			continue
		}
		allChirps = append(allChirps, Chirp{ID: dbChirp.ID, Body: dbChirp.Body, CreatedAt: dbChirp.CreatedAt, UpdatedAt: dbChirp.CreatedAt, UserID: dbChirp.UserID})
	}
	sortDirection := "asc"
	sortDirectionParam := r.URL.Query().Get("sort")
	if sortDirectionParam == "desc" {
		sortDirection = "desc"
	}

	sort.Slice(allChirps, func(i, j int) bool {
		if sortDirection == "desc" {
			return allChirps[i].CreatedAt.After(allChirps[j].CreatedAt)
		}
		return allChirps[i].CreatedAt.Before(allChirps[j].CreatedAt)
	})

	respondWithJSON(w, http.StatusOK, allChirps)

}
