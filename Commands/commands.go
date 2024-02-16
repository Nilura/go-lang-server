package commands

import (
	"fmt"
	"os"

	"github.com/slack-go/slack"
)

var targetChannelID string

func HandleSlashCommand(command slack.SlashCommand, client *slack.Client) error {

	switch command.Command {
	case "/set-channel":

		return handleSetTargetChannelCommand(command, client)
	case "/reset":

		return handleResetTargetChannelCommand(command, client)
	}
	return nil
}

func handleSetTargetChannelCommand(command slack.SlashCommand, client *slack.Client) error {

	_, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionText("Please enter the target channel ID:", false))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	targetChannelID = command.Text

	os.Setenv("ERROR_CHANNEL_ID", targetChannelID)

	response := fmt.Sprintf("Target channel ID set to %s", targetChannelID)
	_, _, err = client.PostMessage(command.ChannelID, slack.MsgOptionText(response, false))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	// _, _, err = client.PostMessage(command.ChannelID, slack.MsgOptionText("Please enter the keyword to filter:", false))

	// if err != nil {
	// 	return fmt.Errorf("failed to post message: %w", err)
	// }

	confirmationMsg := "Successfully registered"
	_, _, err = client.PostMessage(command.ChannelID, slack.MsgOptionText(confirmationMsg, false))
	if err != nil {
		return fmt.Errorf("failed to post confirmation message: %w", err)
	}
	return nil
}

func handleResetTargetChannelCommand(command slack.SlashCommand, client *slack.Client) error {
	os.Unsetenv("ERROR_CHANNEL_ID")
	targetChannelID = ""

	response := "Target channel ID has been reset"
	_, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionText(response, false))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	return nil
}

// func handleHelloCommand(command slack.SlashCommand, client *slack.Client) error {

// 	attachment := slack.Attachment{}

// 	attachment.Fields = []slack.AttachmentField{
// 		{
// 			Title: "Date",
// 			Value: time.Now().String(),
// 		}, {
// 			Title: "Initializer",
// 			Value: command.UserName,
// 		},
// 	}

// 	attachment.Text = fmt.Sprintf("Hello %s", command.Text)
// 	attachment.Color = "#4af030"

// 	_, _, err := client.PostMessage(targetChannelID, slack.MsgOptionAttachments(attachment))
// 	if err != nil {
// 		return fmt.Errorf("failed to post message: %w", err)
// 	}

// 	return nil
// }
