package main

import (
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
	
	serv := &http.Server{
		Addr: ":"+port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(serv.ListenAndServe())
}