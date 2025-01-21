package handlers

import (
	"github.com/mlvieira/nsfwdetection/internal/repositories"
	"github.com/mlvieira/nsfwdetection/internal/websockets"
)

type Handlers struct {
	Repositories *repositories.Repositories
	Hub          *websockets.Hub
}

func NewHandlers(repositories *repositories.Repositories, hub *websockets.Hub) *Handlers {
	return &Handlers{
		Repositories: repositories,
		Hub:          hub,
	}
}
