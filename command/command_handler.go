package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler struct {
	session            *discordgo.Session
	guild              *discordgo.Guild
	commands           map[string]*Command
	registeredCommands []*discordgo.ApplicationCommand
}

func NewCommandHandler(session *discordgo.Session, guild *discordgo.Guild) (*CommandHandler, error) {
	return &CommandHandler{
		session:  session,
		guild:    guild,
		commands: make(map[string]*Command),
	}, nil
}

// Register registers commands
// Make sure to register commands before starting the service
func (c *CommandHandler) Register(commands ...*Command) {
	for _, cmd := range commands {
		c.commands[cmd.Name] = cmd
	}
}

func (c *CommandHandler) Init() error {
	for _, cmd := range c.commands {
		value, err := c.session.ApplicationCommandCreate(c.session.State.Application.ID, c.guild.ID, cmd.ApplicationCommand())
		if err != nil {
			return err
		}
		c.registeredCommands = append(c.registeredCommands, value)
	}
	return nil
}

func (c *CommandHandler) Deinit() error {
	for _, cmd := range c.registeredCommands {
		if err := c.session.ApplicationCommandDelete(c.session.State.User.ID, "", cmd.ID); err != nil {
			return err
		}
	}
	return nil
}

func (c *CommandHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	command, ok := c.commands[i.ApplicationCommandData().Name]
	if !ok {
		return nil, fmt.Errorf("command `%s` not found", i.ApplicationCommandData().Name)
	}
	return command.Handler(s, i)
}
