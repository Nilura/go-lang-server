package view

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetUserName(accessToken, userID string) (string, error) {

	req, err := http.NewRequest("GET", "https://slack.com/api/users.info?user="+userID, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	var response map[string]interface{}
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return "", err
	}
	if !response["ok"].(bool) {
		return "", fmt.Errorf("error from Slack API: %s", response["error"].(string))
	}

	user := response["user"].(map[string]interface{})
	return user["real_name"].(string), nil
}
