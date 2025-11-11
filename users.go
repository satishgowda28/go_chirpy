package main

import (
	"database/sql"
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

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpadatedAt   time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpRed   bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type newUserData struct {
		Email    string `json:"email"`
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
		Email:          newUser.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		errorStr := fmt.Sprintf("Issues in creating users: %s", err)
		respondWithError(w, 400, errorStr)
		return
	}

	respondWithJSON(w, 201, response{User: User{
		ID:         user.ID,
		CreatedAt:  user.CreatedAt,
		UpadatedAt: user.UpdatedAt,
		Email:      user.Email,
		IsChirpRed: user.IsChirpyRed.Bool,
	}})
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {

	type params struct {
		Email           string `json:"Email"`
		Password        string `json:"password"`
		ExpireInSeconds int    `json:"expires_in_seconds"`
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
	if !mached {
		respondWithError(w, http.StatusUnauthorized, "Use is not authorized")
		return
	}

	expireIn := time.Hour
	if loginParams.ExpireInSeconds != 0 && loginParams.ExpireInSeconds <= 3600 {
		expireIn = time.Duration(loginParams.ExpireInSeconds) * time.Second
	}
	jwtToken, err := auth.MakeJWT(user.ID, cfg.secret, expireIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Issue with toekn generation")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rfToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpadatedAt:   user.UpdatedAt,
		Email:        user.Email,
		Token:        jwtToken,
		RefreshToken: rfToken.Token,
		IsChirpRed:   user.IsChirpyRed.Bool,
	})
}

func (cfg *apiConfig) handleUpdateCreds(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		Email string `json:"email"`
	}

	newCreds := params{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newCreds)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error in decoding data")
		return
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

	user, err := cfg.db.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "user with this email is not present")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hashedPassword, err := auth.HashPassord(newCreds.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	user, err = cfg.db.UpdateUserCreds(r.Context(), database.UpdateUserCredsParams{
		Email:          newCreds.Email,
		HashedPassword: hashedPassword,
		ID:             user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Email: user.Email,
	})
}

func (cfg *apiConfig) handleUserChirpSubscription(w http.ResponseWriter, r *http.Request) {

	const UPGRADE = "user.upgraded"

	type payLoad struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	eventData := payLoad{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&eventData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if strings.ToLower(eventData.Event) != UPGRADE {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(eventData.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	userData, err := cfg.db.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "user not found")
		}
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	if userData.IsChirpyRed.Bool {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userData, err = cfg.db.UpdateUserChirpSubscription(r.Context(), userData.ID)
	fmt.Println("user data is_chirpy_red:", userData.IsChirpyRed.Bool)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, User{
		ID:         userData.ID,
		IsChirpRed: userData.IsChirpyRed.Bool,
	})

}
