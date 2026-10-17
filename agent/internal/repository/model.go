package repository

import "github.com/ZhengHeOwo/agent_noah/agent/internal/model"

type MemoryStore interface {
	SaveMessage(message model.Message) error
	GetMessages(prompt string) ([]model.Message, error)
}
