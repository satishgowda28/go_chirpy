package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/satishgowda28/go_chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	secret         string
	polkaApiKey    string
}

func handleFileServe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte("Welcome to Chirpy"))
	// http.StripPrefix("/app", http.FileServer(http.Dir("./")))
}

func main() {

	godotenv.Load()

	port := "8080"
	dbURL := os.Getenv("DB_URL")
	tokenSec := os.Getenv("TOKEN_SEC")
	polkaKey := os.Getenv("POLKA_KEY")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error connection db: %s", err)
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()

	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
		db:             dbQueries,
		secret:         tokenSec,
		polkaApiKey:    polkaKey,
	}
	mux.Handle("/app/assets/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./")))))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.HandlerFunc(handleFileServe)))

	mux.HandleFunc("GET /api/healthz", checkReadiness)

	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)

	// mux.HandleFunc("POST /api/validate_chirp", handleChirpValidation)

	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	mux.HandleFunc("PUT /api/users", apiCfg.handleUpdateCreds)
	mux.HandleFunc("POST /api/login", apiCfg.handleLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handleResetToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.handleRevokeRefreshToken)
	mux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.getAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handleDeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handleUserChirpSubscription)

	/* TESTING ---- Pathvalue */
	/* mux.HandleFunc("GET /api/test/{id}/{name}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		name := r.PathValue("name")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("id: %s, name: %s", id, strings.Trim(name," "))))
	}) */

	serv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(serv.ListenAndServe())
}
