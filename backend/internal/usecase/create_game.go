package usecase

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rafawastaken/ai-hunger-games/internal/domain"
	"github.com/rafawastaken/ai-hunger-games/internal/repository"
)

type CreateGameInput struct {
	NumAgents  int
	MaxStrikes int
}

type CreateGameOutput struct {
	Game *domain.Game
}

type CreateGameUseCase struct {
	gameRepo repository.GameRepository
}

func NewCreateGameUseCase(repo repository.GameRepository) *CreateGameUseCase {
	return &CreateGameUseCase{gameRepo: repo}
}

func (uc *CreateGameUseCase) Execute(input CreateGameInput) (*CreateGameOutput, error) {
	if input.NumAgents <= 0 {
		input.NumAgents = 4
	}
	if input.MaxStrikes <= 0 {
		input.MaxStrikes = 2
	}

	game := &domain.Game{
		ID:         uuid.NewString(),
		MaxStrikes: input.MaxStrikes,
		Status:     domain.GameStatusWaiting,
	}

	agents := make([]*domain.Agent, 0, input.NumAgents)
	for i := 0; i < input.NumAgents; i++ {
		a := &domain.Agent{
			ID:   fmt.Sprintf("agent-%d", i+1),
			Name: fmt.Sprintf("Agent %d", i+1),
		}
		agents = append(agents, a)
	}
	game.Agents = agents

	if err := uc.gameRepo.Create(game); err != nil {
		return nil, err
	}

	return &CreateGameOutput{Game: game}, nil
}
