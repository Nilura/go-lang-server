package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	commands "webhook/Commands"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
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
}

var errorChannelID string

func main() {
	godotenv.Load(".env")

	http.HandleFunc("/slack/oauth/callback", handleOAuthCallback)

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
	appToken := os.Getenv("SLACK_APP_TOKEN")

	webhookURL, accessTokenResponse := getOAuthAccessToken(code)
	fmt.Fprintf(w, "Webhook: %s", webhookURL)
	fmt.Fprintf(w, "Token: %s", code)
	api := slack.New(accessTokenResponse.AccessToken, slack.OptionDebug(true))

	client := slack.New(accessTokenResponse.AccessToken, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))

	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	var shouldProcessMessages = true

	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {

		for {
			select {

			case <-ctx.Done():
				log.Println("Shutting down socketmode listener")
				return
			case event := <-socketClient.Events:

				switch event.Type {

				case socketmode.EventTypeEventsAPI:

					eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)

					if !ok {
						log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
						continue
					}

					socketClient.Ack(*event.Request)

					if shouldProcessMessages {

						err := handleInteractiveCallback(eventsAPIEvent, client, errorChannelID)

						if err != nil {
							log.Fatal(err)
						}

					}

					// Handle Slash Commands
				case socketmode.EventTypeSlashCommand:

					command, ok := event.Data.(slack.SlashCommand)
					if !ok {
						log.Printf("Could not type cast the message to a SlashCommand: %v\n", command)
						continue
					}

					socketClient.Ack(*event.Request)

					err := commands.HandleSlashCommand(command, client)
					if err != nil {
						log.Fatal(err)
					}

				}

			}
		}
	}(ctx, client, socketClient)

	socketClient.Run()

	channelID, timestamp, err := api.PostMessage(accessTokenResponse.IncomingWebhook.ChannelID, slack.MsgOptionText(webhookURL, false))
	if err != nil {
		fmt.Printf("Error posting message to channel: %s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)

	message := fmt.Sprintf("Webhook URL: %s", webhookURL)

	_, _, err = api.PostMessage(accessTokenResponse.UserID, slack.MsgOptionText(message, false))
	if err != nil {
		fmt.Printf("Error posting message to Message tab: %s\n", err)
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

func handleInteractiveCallback(interactiveCallback slackevents.EventsAPIEvent, client *slack.Client, errorChannelID string) error {

	switch interactiveCallback.Type {

	case slackevents.CallbackEvent:

		innerEvent := interactiveCallback.InnerEvent

		switch ev := innerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			err := handleMessageEvent(ev, client, errorChannelID)
			if err != nil {
				return err
			}
		}

	default:
		return errors.New("unsupported event type")
	}
	return nil

}

func handleMessageEvent(event *slackevents.MessageEvent, client *slack.Client, errorChannelID string) error {
	botInfo, err := client.AuthTest()
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}

	if event.User == botInfo.UserID {
		log.Printf("Ignoring own bot message: %s", event.Text)
		return nil
	}
	text := strings.ToLower(event.Text)

	if strings.Contains(text, "error") {

		_, _, _ = client.PostMessage(errorChannelID, slack.MsgOptionText(event.Text, false))

		return nil
	}

	return nil
}

func reloadConfiguration() {
	for {

		time.Sleep(5 * time.Second)

		newTargetChannelID := os.Getenv("ERROR_CHANNEL_ID")

		if newTargetChannelID != errorChannelID {
			errorChannelID = newTargetChannelID

		}
	}
}
