package main

import (
	"go-lang-server/auth"
	commands "go-lang-server/commands"
	events "go-lang-server/events"

	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	http.HandleFunc("/slack/oauth/callback", auth.HandleOAuthCallback)
	http.HandleFunc("/slack/command", commands.HandleSlashCommand)
	http.HandleFunc("/slack/reset", commands.HandleResetCommand)
	http.HandleFunc("/slack/keyword", commands.HandleKeywordCommand)
	http.HandleFunc("/slack/events", events.HandleEvent)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8082"
	}

	log.Printf("Server listening on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
