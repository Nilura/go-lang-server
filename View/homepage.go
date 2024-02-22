package view

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func PublishHomeView(accessToken string, userID string, payload map[string]interface{}) error {

	payload["user_id"] = userID
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/views.publish", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	var response map[string]interface{}
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return err
	}
	if !response["ok"].(bool) {
		return fmt.Errorf("error from Slack API: %s", response["error"].(string))
	}

	return nil
}
