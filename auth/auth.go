package auth

import (
	"encoding/json"
	"fmt"
	"go-lang-server/view"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/slack-go/slack"
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

var AccessTokenMap = make(map[string]string)
var UserIdMap = make(map[string]string)

func HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	webhookURL, accessTokenResponse := getOAuthAccessToken(code)
	fmt.Fprintf(w, "The Randoli Slack App was installed successfully\nPlease copy the following webhook url when you create the integration (This is also available from the Slack Apps Home Page): %s", webhookURL)

	api := slack.New(accessTokenResponse.AccessToken, slack.OptionDebug(true))

	token := accessTokenResponse.Bot.Bot_access_token

	AccessTokenMap[accessTokenResponse.Bot.Bot_user_id] = accessTokenResponse.AccessToken
	UserIdMap[accessTokenResponse.Bot.Bot_user_id] = accessTokenResponse.UserID

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
