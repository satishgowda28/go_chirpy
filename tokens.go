package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/satishgowda28/go_chirpy/internal/auth"
)

func(cfg *apiConfig) handleResetToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	rfToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
	}

	tokenData, err := cfg.db.GetUserFromRefreshToken(r.Context(), rfToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "UnAuthorized token")
			return
		}
	}
	now := time.Now().UTC()
	if tokenData.ExpiresAt.Time.Before(now) {
		respondWithError(w, http.StatusUnauthorized, "Token is expired")
		return
	}
	
	if tokenData.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Token is loggedout")
		return
	}
	
	jwtToken, err := auth.MakeJWT(tokenData.UserID, cfg.secret, time.Hour)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Issues in creating a authToken")
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: jwtToken,
	})

}

func(cfg *apiConfig) handleRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {

	rfToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
	}
	
	tokenData, err := cfg.db.GetUserFromRefreshToken(r.Context(), rfToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "UnAuthorized token")
			return
		}
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), tokenData.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Issue in revoking the token")
		return
	}

	w.WriteHeader(http.StatusNoContent)

}