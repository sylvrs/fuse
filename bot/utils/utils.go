package utils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func GetMessagesFromUser(s *discordgo.Session, u *discordgo.User, c *discordgo.Channel, limit uint) ([]*discordgo.Message, error) {
	raw, err := s.ChannelMessages(c.ID, int(limit), "", "", "")
	if err != nil {
		return nil, err
	}
	messages := make([]*discordgo.Message, 0)
	for _, m := range raw {
		if m.Author.ID == u.ID {
			messages = append(messages, m)
		}
	}
	return messages, nil
}

func GetLatestMessageFromUser(s *discordgo.Session, u *discordgo.User, c *discordgo.Channel) (*discordgo.Message, error) {
	messages, err := GetMessagesFromUser(s, u, c, 1)
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return nil, nil
	}
	return messages[0], nil
}

func Pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func InArray[T comparable](array []T, value T) bool {
	for _, current := range array {
		if current == value {
			return true
		}
	}
	return false
}

func RemoveFromArray[T comparable](array []T, value T) ([]T, error) {
	for i, current := range array {
		if current == value {
			return append(array[:i], array[i+1:]...), nil
		}
	}
	return nil, fmt.Errorf("element not found")
}

// IfElse is a function that returns `a` if `condition` is true, otherwise it returns `b`
func IfElse[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

// Map returns a new array containing the results of applying `mapper` to each element of `array`
func Map[T any, U any](array []T, mapper func(T) U) []U {
	result := make([]U, len(array))
	for i, current := range array {
		result[i] = mapper(current)
	}
	return result
}

// Filter returns a new array containing only the elements of `array` for which `predicate` returns true
func Filter[T any](array []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, current := range array {
		if predicate(current) {
			result = append(result, current)
		}
	}
	return result
}

// Contains returns true if `array` contains `value`
func Contains[T comparable](array []T, value T) bool {
	for _, current := range array {
		if current == value {
			return true
		}
	}
	return false
}

// Remove removes the first instance of `value` from `array`
func Remove[T comparable](array []T, value T) ([]T, error) {
	for i, current := range array {
		if current == value {
			return append(array[:i], array[i+1:]...), nil
		}
	}
	return nil, fmt.Errorf("element not found")
}

func IntPtr[T int64](i T) *T {
	return &i
}
