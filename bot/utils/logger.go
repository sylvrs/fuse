package utils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	log "github.com/inconshreveable/log15"
	"golang.org/x/exp/slices"
)

var (
	// The keys to strip from the log output when logging to Discord
	excludedLoggingKeys = []string{"fn"}
)

const (
	ColorPrimary = 0xB4DDB4
	ColorSuccess = 0x00FF00
	ColorError   = 0xFF0000
	ColorWarning = 0xFFFF00
	ColorInfo    = 0x0000FF
	ColorNeutral = 0x808080
)

func embedOptionsByLevel(lvl log.Lvl) (string, int) {
	switch lvl {
	case log.LvlInfo:
		return "Info", ColorInfo
	case log.LvlWarn:
		return "Warn", ColorWarning
	case log.LvlError:
		return "Error", ColorError
	case log.LvlCrit:
		return "Critical", ColorError
	default:
		return "Unknown", ColorNeutral
	}
}

func NewDiscordLogHandler(s *discordgo.Session, guildId, channelId string) log.Handler {
	return log.FuncHandler(func(r *log.Record) error {
		title, color := embedOptionsByLevel(r.Lvl)

		fields := make([]*discordgo.MessageEmbedField, 0)

		// iterate through context and add fields (2 at a time)
		if (len(r.Ctx) % 2) == 0 {
			for i := 0; i < len(r.Ctx); i += 2 {
				if key, ok := r.Ctx[i].(string); ok && slices.Contains(excludedLoggingKeys, key) {
					continue
				}
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:  fmt.Sprintf("%v", r.Ctx[i]),
					Value: fmt.Sprintf("%v", r.Ctx[i+1]),
				})
			}
		}

		_, _ = s.ChannelMessageSendEmbed(channelId, &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Logger | %s", title),
			Description: r.Msg,
			Color:       color,
			Fields:      fields,
		})
		return nil
	})
}
