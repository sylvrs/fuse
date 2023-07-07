package modal

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sylvrs/fuse/bot/utils"
)

type ModalHandlerFunc func(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error)
type Modal struct {
	Id         string
	Title      string
	Components []discordgo.MessageComponent
	Handler    ModalHandlerFunc
}

func NewModal(id, title string, components []discordgo.MessageComponent, handler ModalHandlerFunc) *Modal {
	return &Modal{
		Id:         id,
		Title:      title,
		Components: components,
		Handler:    handler,
	}
}

func NewTextModal(id, title string, inputs []discordgo.TextInput, handler ModalHandlerFunc) *Modal {
	return &Modal{
		Id:    id,
		Title: title,
		Components: utils.Map(inputs, func(input discordgo.TextInput) discordgo.MessageComponent {
			return &discordgo.ActionsRow{Components: []discordgo.MessageComponent{&input}}
		}),
		Handler: handler,
	}
}

// Clone returns a copy of the modal.
// This is useful for creating static modals that can be reused but have different data.
func (m *Modal) Clone() *Modal {
	return &Modal{
		Id:         m.Id,
		Title:      m.Title,
		Components: m.Components,
		Handler:    m.Handler,
	}
}

func (m *Modal) ModalData(suffix string) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		CustomID:   fmt.Sprintf("%s-%s", m.Id, suffix),
		Title:      m.Title,
		Components: m.Components,
	}
}
