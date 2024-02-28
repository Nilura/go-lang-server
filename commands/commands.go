package commands

import (
	"fmt"
	"net/http"
)

type KeyValue struct {
	ChannelID string
	Keyword   string
}

var ChannelMap = make(map[string]string)
var KeyMap = make(map[string]KeyValue)

func HandleSlashCommand(w http.ResponseWriter, r *http.Request) {

	err := handleSlashCommandFromHTTP(w, r)
	if err != nil {
		http.Error(w, "Error handling slash command: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleResetCommand(w http.ResponseWriter, r *http.Request) {

	err := handleResetCommandFromHTTP(w, r)
	if err != nil {
		http.Error(w, "Error handling slash command: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleKeywordCommand(w http.ResponseWriter, r *http.Request) {

	err := handleKeywordCommandFromHTTP(w, r)
	if err != nil {
		http.Error(w, "Error handling slash command: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleSlashCommandFromHTTP(w http.ResponseWriter, r *http.Request) error {

	err := r.ParseForm()
	if err != nil {
		return err
	}
	fmt.Println("==========")
	fmt.Println(r)
	text := r.Form.Get("text")
	userID := r.Form.Get("user_id")
	fmt.Println("userId:%s" + userID)
	ChannelMap[userID] = text

	responseText := fmt.Sprintf("Successfully registered the channel %s", text)

	_, err = w.Write([]byte(responseText))
	if err != nil {
		return err
	}

	return nil
}

func handleKeywordCommandFromHTTP(w http.ResponseWriter, r *http.Request) error {

	err := r.ParseForm()
	if err != nil {
		return err
	}
	text := r.Form.Get("text")
	userID := r.Form.Get("user_id")
	channelID := ChannelMap[userID]

	newValue := KeyValue{
		ChannelID: channelID,
		Keyword:   text,
	}

	KeyMap[userID] = newValue
	fmt.Println("&&&&&&&&& %s", KeyMap)
	responseText := "The keyword has been set to the target channel."

	_, err = w.Write([]byte(responseText))
	if err != nil {
		return err
	}

	return nil
}

func handleResetCommandFromHTTP(w http.ResponseWriter, r *http.Request) error {

	err := r.ParseForm()
	if err != nil {
		return err
	}
	//userID := r.Form.Get("user_id")
	ChannelMap = make(map[string]string)
	KeyMap = make(map[string]KeyValue)
	fmt.Println("###### %s", KeyMap)
	fmt.Println("@@@@@@ %s", ChannelMap)

	responseText := "Target channel ID and Keyword has been reset "
	fmt.Println(responseText)
	_, err = w.Write([]byte(responseText))
	if err != nil {
		return err
	}

	return nil
}
