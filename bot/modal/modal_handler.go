package modal

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type ModalHandler struct {
	session       *discordgo.Session
	guild         *discordgo.Guild
	pendingModals map[string]*Modal
}

func NewModalHandler(session *discordgo.Session, guild *discordgo.Guild) *ModalHandler {
	return &ModalHandler{
		session:       session,
		guild:         guild,
		pendingModals: make(map[string]*Modal),
	}
}

// Send is used to send a modal to a user based on an interaction
// This will return an interaction response that can be used to send the modal
func (h *ModalHandler) Send(i *discordgo.InteractionCreate, m *Modal) (*discordgo.InteractionResponse, error) {
	data := m.ModalData(i.Member.User.ID)
	h.pendingModals[data.CustomID] = m.Clone()
	// return the interaction response
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: data,
	}, nil
}

func (h *ModalHandler) Handle(i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	// get the modal from the pending modals
	m, ok := h.pendingModals[i.ModalSubmitData().CustomID]
	if !ok {
		return nil, fmt.Errorf("modal '%s' not found", i.ModalSubmitData().CustomID)
	}
	// delete the modal from the pending modals
	delete(h.pendingModals, i.ModalSubmitData().CustomID)
	// call the modal handler
	return m.Handler(h.session, i)
}
