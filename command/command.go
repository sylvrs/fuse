package command

import "github.com/bwmarrin/discordgo"

type Command struct {
	Name               string
	Description        string
	DefaultPermissions *int64
	Options            []*discordgo.ApplicationCommandOption
	Handler            func(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error)
}

func (c *Command) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:                     c.Name,
		Description:              c.Description,
		DefaultMemberPermissions: c.DefaultPermissions,
		Options:                  c.Options,
	}
}
