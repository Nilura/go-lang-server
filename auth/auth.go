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
	OK         bool   `json:"ok"`
	AppID      string `json:"app_id"`
	AuthedUser struct {
		ID          string `json:"id"`
		Scope       string `json:"scope"`
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	} `json:"authed_user"`
	Scope          string `json:"scope"`
	TokenType      string `json:"token_type"`
	BotAccessToken string `json:"access_token"`
	BotUserID      string `json:"bot_user_id"`
	Team           struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	Enterprise          interface{} `json:"enterprise"`
	IsEnterpriseInstall bool        `json:"is_enterprise_install"`
	IncomingWebhook     struct {
		Channel          string `json:"channel"`
		ChannelID        string `json:"channel_id"`
		ConfigurationURL string `json:"configuration_url"`
		URL              string `json:"url"`
	} `json:"incoming_webhook"`
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

	api := slack.New(accessTokenResponse.BotAccessToken, slack.OptionDebug(true))

	token := accessTokenResponse.BotAccessToken

	
	AccessTokenMap[accessTokenResponse.BotUserID] = accessTokenResponse.AuthedUser.AccessToken
	UserIdMap[accessTokenResponse.BotUserID] = accessTokenResponse.AuthedUser.ID
	
	message := fmt.Sprintf("Webhook URL: %s", webhookURL)
        fmt.Println("=================")
	fmt.Println(accessTokenResponse.AuthedUser.ID)
        fmt.Println(accessTokenResponse.BotUserID)
        fmt.Println("=================")

	_, _, err := api.PostMessage("", slack.MsgOptionText(message, false))
	if err != nil {
		fmt.Printf("Error posting message to Message tab: %s\n", err)
		return
	}

	userName, err := view.GetUserName(token, accessTokenResponse.AuthedUser.ID)
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

        er := view.PublishHomeView(token, accessTokenResponse.AuthedUser.ID, payload)
	if er != nil {
		fmt.Println("Error publishing home view:", er)
		return
	}

}

func getOAuthAccessToken(code string) (string, AccessTokenResponse) {
	clientID := os.Getenv("SLACK_CLIENT_ID")
	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")
	redirectURI := os.Getenv("SLACK_REDIRECT_URI")

	url := fmt.Sprintf("https://slack.com/api/oauth.v2.access?client_id=%s&client_secret=%s&code=%s&redirect_uri=%s", clientID, clientSecret, code, redirectURI)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

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
