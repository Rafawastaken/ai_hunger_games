package usecase

import (
	"context"
	"fmt"

	"github.com/rafawastaken/ai-hunger-games/internal/domain"
	"github.com/rafawastaken/ai-hunger-games/internal/repository"
	"github.com/rafawastaken/ai-hunger-games/internal/service"
)

type PlayRoundInput struct {
	GameID   string
	Question string
}

type PlayRoundOutput struct {
	Game  *domain.Game
	Round *domain.Round
}

type PlayRoundUseCase struct {
	gameRepo    repository.GameRepository
	groq        service.GroqService
	debateTurns int
}

func NewPlayRoundUseCase(repo repository.GameRepository, groq service.GroqService) *PlayRoundUseCase {
	return &PlayRoundUseCase{
		gameRepo:    repo,
		groq:        groq,
		debateTurns: 2, // reduzido para economizar tokens
	}
}

func (uc *PlayRoundUseCase) Execute(ctx context.Context, input PlayRoundInput) (*PlayRoundOutput, error) {
	game, err := uc.gameRepo.Get(input.GameID)
	if err != nil {
		return nil, err
	}
	if game.Status == domain.GameStatusFinished {
		return nil, fmt.Errorf("game already finished")
	}
	if input.Question == "" {
		return nil, fmt.Errorf("question is required")
	}

	activeAgents := game.ActiveAgents()
	if len(activeAgents) == 0 {
		return nil, fmt.Errorf("no active agents in game")
	}

	round := &domain.Round{
		Index:    game.NextRoundIndex(),
		Question: input.Question,
	}

	// 1) Respostas iniciais
	for _, agent := range activeAgents {
		text, err := uc.groq.GenerateAnswer(ctx, game, agent, input.Question)
		if err != nil {
			return nil, err
		}
		round.Answers = append(round.Answers, domain.Answer{
			AgentID: agent.ID,
			Text:    text,
		})
	}

	// 2) Debate
	for turn := 1; turn <= uc.debateTurns; turn++ {
		for _, agent := range activeAgents {
			msg, err := uc.groq.GenerateDebateMessage(ctx, game, round, agent)
			if err != nil {
				return nil, err
			}
			round.Debate = append(round.Debate, domain.DebateMessage{
				AgentID: agent.ID,
				Turn:    turn,
				Text:    msg,
			})
		}
	}

	// 3) Votação
	votesCount := make(map[string]int)

	// inicializar todos a 0 para zeros também contarem como pior score
	for _, agent := range activeAgents {
		votesCount[agent.ID] = 0
	}

	for _, agent := range activeAgents {
		targetID, justification, err := uc.groq.GenerateVote(ctx, game, round, agent)
		if err != nil {
			return nil, err
		}
		round.Votes = append(round.Votes, domain.Vote{
			VoterID:       agent.ID,
			TargetID:      targetID,
			Justification: justification,
		})
		if _, ok := votesCount[targetID]; ok {
			votesCount[targetID]++
		}
	}

	// 4) Determinar quem levou strike (menos votos)
	if len(activeAgents) > 0 {
		minVotes := -1
		for _, agent := range activeAgents {
			count := votesCount[agent.ID]
			if minVotes == -1 || count < minVotes {
				minVotes = count
			}
		}

		for _, agent := range activeAgents {
			if votesCount[agent.ID] == minVotes {
				agent.Strikes++
				if agent.Strikes >= game.MaxStrikes {
					agent.Eliminated = true
					round.Eliminated = append(round.Eliminated, agent.ID)
				}
			}
		}
	}

	// 5) Atualizar estado do jogo
	game.Rounds = append(game.Rounds, round)

	if len(game.ActiveAgents()) <= 1 {
		game.Status = domain.GameStatusFinished
	} else {
		game.Status = domain.GameStatusRunning
	}

	if err := uc.gameRepo.Update(game); err != nil {
		return nil, err
	}

	return &PlayRoundOutput{
		Game:  game,
		Round: round,
	}, nil
}
