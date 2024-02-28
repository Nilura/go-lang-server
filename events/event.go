package events

import (
	"encoding/json"
	"fmt"
	auth "go-lang-server/auth"
	"go-lang-server/commands"
	"go-lang-server/view"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
		text := r.Form.Get("metadata")
	fmt.Println(text)
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

		attachments, ok := eventData["attachments"].([]interface{})
		if !ok || len(attachments) == 0 {
			fmt.Println("No attachments found")
			return
		}

	     
		
		// // Assuming there's only one attachment for simplicity
		 attachment := attachments[0].(map[string]interface{})
                 fmt.Println("Extracted text:", attachment)
		blocks, ok := attachment["blocks"].([]interface{})
		if !ok || len(blocks) == 0 {
			fmt.Println("No blocks found in the attachment")
			return
		}
             
		
	   
		block := blocks[0].(map[string]interface{})
	
		fields := block["fields"].([]interface{})
		fmt.Println("Blocks text:", fields[0])
		if len(fields) > 0 {
		    firstField := fields[0].(map[string]interface{})
		    text := firstField["text"].(string)
		    fmt.Println("Text of the first field:", text)
		} else {
		    fmt.Println("No fields found in the block")
		}
		userData, ok := payload["authorizations"].([]interface{})
		if !ok || len(userData) == 0 {
			http.Error(w, "User data not found", http.StatusBadRequest)
			return
		}

		userAuthorization, ok := userData[0].(map[string]interface{})
		if !ok {
			http.Error(w, "User authorization data not found", http.StatusBadRequest)
			return
		}

		userID, ok := userAuthorization["user_id"].(string)
		if !ok {
			http.Error(w, "User ID not found", http.StatusBadRequest)
			return
		}
		fmt.Println("Received UserId:", userID)

		accessToken, found := auth.AccessTokenMap[userID]
		user_id, ok := auth.UserIdMap[userID]

		if !ok {
			http.Error(w, "User ID not found", http.StatusBadRequest)
			return
		}

		fmt.Println("accessToken:", accessToken)
		fmt.Println("Map:", auth.AccessTokenMap)
		if !found {
			http.Error(w, "Access token not found for bot ID", http.StatusInternalServerError)
			return
		}
		switch eventType {
		case "message":

			postEventToChannel(accessToken, eventData, user_id)
		case "reaction_added":

		default:
			fmt.Printf("Unsupported event type: %s\n", eventType)
		}

	default:
		http.Error(w, "Unsupported event type", http.StatusBadRequest)
	}
}

var messageCache = make(map[string]bool)

func postEventToChannel(token string, eventData map[string]interface{}, userId string) error {

	text := eventData["text"].(string)
	keyValue := commands.KeyMap[userId]
	fmt.Println("Received event data:", eventData)

	if strings.Contains(text, keyValue.Keyword) {

		if messageCache[text] {
			fmt.Printf("Message with text '%s' has already been posted\n", text)
			return nil
		}

		fmt.Println("Posting message to channel...")
		err := view.PublishMsg(token, commands.ChannelMap[userId], eventData)
		if err != nil {
			return err
		}

		messageCache[text] = true
		fmt.Printf("Message with text '%s' has been successfully posted\n", text)
	}

	return nil
}
