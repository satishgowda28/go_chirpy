package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func handleFileServe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte("Welcome to Chirpy"))
	// http.StripPrefix("/app", http.FileServer(http.Dir("./")))
}

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


func main () {

	port := "8080"

	mux := http.NewServeMux()
	
	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
	}
	mux.Handle("/app/assets/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./")))))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.HandlerFunc(handleFileServe)))
	
	mux.HandleFunc("GET /api/healthz", checkReadiness)
	
	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)

	mux.HandleFunc("POST /api/validate_chirp", handleChirpValidation)

	
	serv := &http.Server{
		Addr: ":"+port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(serv.ListenAndServe())
}