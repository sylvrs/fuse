package utils

import (
	"github.com/bwmarrin/discordgo"
)

// UpdateMessages compares the current version of a message to the version that was sent and updates it if they are different
func UpdateMessage(session *discordgo.Session, current *discordgo.MessageSend, sent *discordgo.Message) error {
	if MessagesMatch(current, sent) {
		return nil
	}

	_, err := session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         sent.ID,
		Channel:    sent.ChannelID,
		Content:    &current.Content,
		Embeds:     &current.Embeds,
		Components: &current.Components,
	})
	return err
}

func MessagesMatch(current *discordgo.MessageSend, sent *discordgo.Message) bool {
	clonedCurrent := cloneMessage(current).(*discordgo.MessageSend)
	clonedSent := cloneMessage(sent).(*discordgo.Message)
	currentMarshaled, _ := discordgo.Marshal(clonedCurrent)
	sentMarshaled, _ := discordgo.Marshal(&discordgo.MessageSend{
		Content:    clonedSent.Content,
		Embeds:     clonedSent.Embeds,
		Components: clonedSent.Components,
	})
	return string(currentMarshaled) == string(sentMarshaled)
}

func cloneMessage(m any) any {
	switch m := m.(type) {
	case *discordgo.Message:
		cloned := *m
		for _, embed := range cloned.Embeds {
			if embed.Type == "" {
				embed.Type = "rich"
			}
		}
		return &cloned
	case *discordgo.MessageSend:
		cloned := *m
		for _, embed := range cloned.Embeds {
			if embed.Type == "" {
				embed.Type = "rich"
			}
		}
		return &cloned
	default:
		return nil
	}
}

func ErrorAsEmbed(message string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Error",
		Description: message,
		Color:       ColorError,
	}
}

func SuccessAsEmbed(description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Success",
		Description: description,
		Color:       ColorSuccess,
	}
}

func InfoAsEmbed(description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Info",
		Description: description,
		Color:       ColorInfo,
	}
}
