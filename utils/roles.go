package utils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// fetchGuildRoles returns the guild's roles from in-memory (if possible)
// If it can't and/or it isn't cached, it will fall back to the REST API.
func fetchGuildRoles(s *discordgo.Session, guildID string) ([]*discordgo.Role, error) {
	if guild, err := s.State.Guild(guildID); err == nil {
		return guild.Roles, nil
	}
	return s.GuildRoles(guildID)
}

// GetRoleByName returns the role with the given name or an error if it doesn't exist
func GetRoleByName(s *discordgo.Session, guildID string, name string) (*discordgo.Role, error) {
	roles, err := fetchGuildRoles(s, guildID)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role.Name == name {
			return role, nil
		}
	}
	return nil, fmt.Errorf("role not found")
}

// GetRoleByID returns the role with the given id or an error if it doesn't exist
func GetRoleByID(s *discordgo.Session, guildID string, id string) (*discordgo.Role, error) {
	roles, err := fetchGuildRoles(s, guildID)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role.ID == id {
			return role, nil
		}
	}
	return nil, fmt.Errorf("role not found")
}

// HasRole iterates over the member's roles and checks if the role is present
func HasRole(s *discordgo.Session, member *discordgo.Member, role *discordgo.Role) bool {
	for _, r := range member.Roles {
		if r == role.ID {
			return true
		}
	}
	return false
}

// GetMemberRoles returns the roles of the member in the given guild
func GetMemberRoles(s *discordgo.Session, guildID string, member *discordgo.Member) ([]*discordgo.Role, error) {
	roles, err := fetchGuildRoles(s, guildID)
	if err != nil {
		return nil, err
	}

	memberRoles := make([]*discordgo.Role, 0)
	for _, r := range member.Roles {
		for _, role := range roles {
			if r == role.ID {
				memberRoles = append(memberRoles, role)
			}
		}
	}
	return memberRoles, nil
}