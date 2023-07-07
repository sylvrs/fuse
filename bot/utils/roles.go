package utils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// GetRoleById returns the role with the given name or an error if it doesn't exist
func GetRoleByName(s *discordgo.Session, guildId string, name string) (*discordgo.Role, error) {
	roles, err := s.GuildRoles(guildId)
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

// GetRoleById returns the role with the given id or an error if it doesn't exist
func GetRoleById(s *discordgo.Session, guildId string, id string) (*discordgo.Role, error) {
	roles, err := s.GuildRoles(guildId)
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
func GetMemberRoles(s *discordgo.Session, guildId string, member *discordgo.Member) ([]*discordgo.Role, error) {
	roles, err := s.GuildRoles(guildId)
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
