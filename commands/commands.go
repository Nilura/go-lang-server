package commands

import (
	"fmt"
	"net/http"
	"os"
)

func HandleSlashCommandFromHTTP(w http.ResponseWriter, r *http.Request) error {

	err := r.ParseForm()
	if err != nil {
		return err
	}

	text := r.Form.Get("text")
	//	userID := r.Form.Get("user_id")

	os.Setenv("ERROR_CHANNEL_ID", text)

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
	//userID := r.Form.Get("user_id")

	os.Setenv("KEYWORD", text)

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

	os.Unsetenv("ERROR_CHANNEL_ID")
	os.Unsetenv("KEYWORD")

	responseText := "Target channel ID has been reset "
	fmt.Println(responseText)
	_, err = w.Write([]byte(responseText))
	if err != nil {
		return err
	}

	return nil
}
