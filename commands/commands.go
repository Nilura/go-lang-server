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

func HandleSlashCommandFromHTTP(w http.ResponseWriter, r *http.Request) error {

	err := r.ParseForm()
	if err != nil {
		return err
	}

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

func HandleKeywordCommandFromHTTP(w http.ResponseWriter, r *http.Request) error {

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
	//	os.Setenv("KEYWORD", text)
	fmt.Println("&&&&&&&&& %s", KeyMap)
	responseText := "The keyword has been set to the target channel."

	_, err = w.Write([]byte(responseText))
	if err != nil {
		return err
	}

	return nil
}

func HandleResetCommandFromHTTP(w http.ResponseWriter, r *http.Request) error {

	err := r.ParseForm()
	if err != nil {
		return err
	}
	//userID := r.Form.Get("user_id")
	ChannelMap = make(map[string]string)
	KeyMap = make(map[string]KeyValue)
	fmt.Println("###### %s", KeyMap)
	fmt.Println("@@@@@@ %s", ChannelMap)
	// os.Unsetenv("ERROR_CHANNEL_ID")
	// os.Unsetenv("KEYWORD")

	responseText := "Target channel ID has been reset "
	fmt.Println(responseText)
	_, err = w.Write([]byte(responseText))
	if err != nil {
		return err
	}

	return nil
}
