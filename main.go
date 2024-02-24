package main

import (
	"encoding/json"
	"fmt"
	commands "go-lang-server/commands"
	event "go-lang-server/events"
	view "go-lang-server/view"
	"io/ioutil"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"

	"log"
	"net/http"
	"os"
	"time"
)

type AccessTokenResponse struct {
	OK              bool   `json:"ok"`
	AccessToken     string `json:"access_token"`
	Scope           string `json:"scope"`
	UserID          string `json:"user_id"`
	TeamID          string `json:"team_id"`
	EnterpriseID    string `json:"enterprise_id"`
	TeamName        string `json:"team_name"`
	IncomingWebhook struct {
		Channel          string `json:"channel"`
		ChannelID        string `json:"channel_id"`
		ConfigurationURL string `json:"configuration_url"`
		URL              string `json:"url"`
	} `json:"incoming_webhook"`
	Bot struct {
		Bot_access_token string `json:"bot_access_token"`
		Bot_user_id      string `json:"bot_user_id"`
	} `json:"bot"`
}

var errorChannelID string
var keyword string

var AccessToken string

func main() {
	godotenv.Load(".env")

	http.HandleFunc("/slack/oauth/callback", handleOAuthCallback)
	http.HandleFunc("/slack/command", handleSlashCommand)
	http.HandleFunc("/slack/reset", handleResetCommand)
	http.HandleFunc("/slack/keyword", handleKeywordCommand)
	//http.HandleFunc("/slack/events", event.HandleEvent)
	http.HandleFunc("/slack/events", func(w http.ResponseWriter, r *http.Request) {
		event.HandleEvent(w, r, AccessToken, errorChannelID)
	})

	port := os.Getenv("PORT")

	if port == "" {
		port = "8082"
	}

	go reloadConfiguration()

	log.Printf("Server listening on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}

func handleOAuthCallback(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	webhookURL, accessTokenResponse := getOAuthAccessToken(code)
	fmt.Fprintf(w, "The Randoli Slack App was installed successfully Please copy the following webhook url when you create the integration (This is also available from the Slack Apps Home Page): %s", webhookURL)

	api := slack.New(accessTokenResponse.AccessToken, slack.OptionDebug(true))

	token := accessTokenResponse.Bot.Bot_access_token
	AccessToken = accessTokenResponse.AccessToken
	message := fmt.Sprintf("Webhook URL: %s", webhookURL)

	_, _, err := api.PostMessage(accessTokenResponse.UserID, slack.MsgOptionText(message, false))
	if err != nil {
		fmt.Printf("Error posting message to Message tab: %s\n", err)
		return
	}

	userName, err := view.GetUserName(token, accessTokenResponse.UserID)
	if err != nil {
		fmt.Println("Error getting user name:", err)
		return
	}

	payload := map[string]interface{}{
		"user_id": "U06G7MKG9V1",
		"view": map[string]interface{}{
			"type": "home",
			"blocks": []map[string]interface{}{
				{
					"type": "header",
					"text": map[string]interface{}{
						"type": "plain_text",
						"text": "Welcome to App Insight Bot",
					},
				},
				{
					"type": "section",
					"text": map[string]interface{}{
						"type": "mrkdwn",
						"text": "Hi " + userName + "\n" + message,
					},
				},
			},
		},
	}

	er := view.PublishHomeView(token, accessTokenResponse.UserID, payload)
	if er != nil {
		fmt.Println("Error publishing home view:", er)
		return
	}

}

func handleSlashCommand(w http.ResponseWriter, r *http.Request) {

	err := commands.HandleSlashCommandFromHTTP(w, r)
	if err != nil {
		http.Error(w, "Error handling slash command: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleResetCommand(w http.ResponseWriter, r *http.Request) {

	err := commands.HandleResetCommandFromHTTP(w, r)
	if err != nil {
		http.Error(w, "Error handling slash command: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleKeywordCommand(w http.ResponseWriter, r *http.Request) {

	err := commands.HandleKeywordCommandFromHTTP(w, r)
	if err != nil {
		http.Error(w, "Error handling slash command: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func getOAuthAccessToken(code string) (string, AccessTokenResponse) {
	clientID := os.Getenv("SLACK_CLIENT_ID")
	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")
	redirectURI := os.Getenv("SLACK_REDIRECT_URI")

	url := fmt.Sprintf("https://slack.com/api/oauth.access?client_id=%s&client_secret=%s&code=%s&redirect_uri=%s", clientID, clientSecret, code, redirectURI)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	if err != nil {
		panic(err)
	}

	var accessTokenResponse AccessTokenResponse
	err = json.Unmarshal(body, &accessTokenResponse)
	if err != nil {
		panic(err)
	}

	return accessTokenResponse.IncomingWebhook.URL, accessTokenResponse
}

func reloadConfiguration() {
	for {
		time.Sleep(5 * time.Second)
		newTargetChannelID := os.Getenv("ERROR_CHANNEL_ID")
		newKeyword := os.Getenv("KEYWORD")
		if newTargetChannelID != errorChannelID {
			errorChannelID = newTargetChannelID

		}
		if newKeyword != keyword {
			keyword = newKeyword

		}

		fmt.Println("channelID", errorChannelID)
		fmt.Println("currentKeyword", keyword)
	}
}
