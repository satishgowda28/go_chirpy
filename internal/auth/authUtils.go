package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassord(password string) (string, error) {
	hashedPass, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPass, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	matched, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return matched, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	currentTime := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(currentTime.Add(expiresIn)),
		Subject:   userID.String(),
	})
	ss, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return ss, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claim := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claim, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		// uuid.UUID{} noo noo use uuid.Nil
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, errors.New("token is invalid")
	}
	id, err := uuid.Parse(claim.Subject)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	encodedStr := hex.EncodeToString([]byte(key))
	return encodedStr, nil
}

func GetBearerToken(header http.Header) (string, error) {
	authToken := header.Get("Authorization")
	if authToken == "" {
		return "", errors.New("auth token is empty")
	}
	if !strings.HasPrefix(authToken, "Bearer ") {
		return "", errors.New("invalid auth token")
	}
	token := strings.TrimPrefix(authToken, "Bearer ")
	token = strings.TrimSpace(token)

	if token == "" {
		return "", errors.New("bearer token is empty")

	}
	return token, nil
}

func GetAPIKey(header http.Header) (string, error) {
	apiKey := header.Get("Authorization")
	if apiKey == "" {
		return "", errors.New("api key is empty")
	}
	if !strings.HasPrefix(apiKey, "ApiKey ") {
		return "", errors.New("api key is invalid")
	}

	key := strings.TrimPrefix(apiKey, "ApiKey ")
	key = strings.TrimSpace(key)

	if key == "" {
		return "", errors.New("api token is empty")
	}

	return key, nil
}
