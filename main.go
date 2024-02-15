package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	http.HandleFunc("/slack/oauth/callback", handleOAuthCallback)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Server listening on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Access Code: %s", code)

}
