package event

import (
	"encoding/json"
	"fmt"
	view "go-lang-server/view"
	"io/ioutil"

	"net/http"
	"os"
	"strings"
)

func HandleEvent(w http.ResponseWriter, r *http.Request, AccessToken string, errorChannelID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Error parsing JSON payload", http.StatusBadRequest)
		return
	}

	eventType, ok := payload["type"].(string)
	if !ok {
		http.Error(w, "Event type not found", http.StatusBadRequest)
		return
	}

	switch eventType {
	case "url_verification":

		challenge, ok := payload["challenge"].(string)
		if !ok {
			http.Error(w, "Challenge parameter not found", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, challenge)

	case "event_callback":

		eventData := payload["event"].(map[string]interface{})

		eventType := eventData["type"].(string)

		switch eventType {
		case "message":

			postEventToChannel(AccessToken, eventData, errorChannelID)
		case "reaction_added":

		default:
			fmt.Printf("Unsupported event type: %s\n", eventType)
		}

	default:
		http.Error(w, "Unsupported event type", http.StatusBadRequest)
	}
}

var messageCache = make(map[string]bool)

func postEventToChannel(token string, eventData map[string]interface{}, errorChannelID string) error {
	keyword := os.Getenv("KEYWORD")
	text := eventData["text"].(string)
	//t := eventData["attachments"].([]map[string]interface{})
	fmt.Println("Received event data:", errorChannelID)

	if strings.Contains(text, keyword) {

		if messageCache[text] {
			fmt.Printf("Message with text '%s' has already been posted\n", text)
			return nil
		}

		fmt.Println("Posting message to channel...")
		err := view.PublishMsg(token, errorChannelID, eventData)
		if err != nil {
			return err
		}

		messageCache[text] = true
		fmt.Printf("Message with text '%s' has been successfully posted\n", text)
	}

	return nil
}
