package component

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type ComponentHandlerFunc func(i *discordgo.InteractionCreate, data *discordgo.MessageComponentInteractionData) (*discordgo.InteractionResponse, error)

func CreateCustomId(prefix, interactionId, componentName string) string {
	return strings.Join([]string{prefix, interactionId, componentName}, "-")
}

func ParseCustomId(customId string) (prefix, interactionId, componentName string) {
	split := strings.Split(customId, "-")
	return split[0], split[1], split[2]
}
