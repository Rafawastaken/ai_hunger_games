package repository

import (
	"errors"
	"sync"

	"github.com/rafawastaken/ai-hunger-games/internal/domain"
)

var ErrGameNotFound = errors.New("game not found")

type GameRepository interface {
	Create(game *domain.Game) error
	Update(game *domain.Game) error
	Get(id string) (*domain.Game, error)
	List() ([]*domain.Game, error)
}

type InMemoryGameRepository struct {
	mu    sync.RWMutex
	games map[string]*domain.Game
}

func NewInMemoryGameRepository() *InMemoryGameRepository {
	return &InMemoryGameRepository{
		games: make(map[string]*domain.Game),
	}
}

func (r *InMemoryGameRepository) Create(game *domain.Game) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.games[game.ID] = game
	return nil
}

func (r *InMemoryGameRepository) Update(game *domain.Game) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.games[game.ID]; !ok {
		return ErrGameNotFound
	}
	r.games[game.ID] = game
	return nil
}

func (r *InMemoryGameRepository) Get(id string) (*domain.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	game, ok := r.games[id]
	if !ok {
		return nil, ErrGameNotFound
	}
	return game, nil
}

func (r *InMemoryGameRepository) List() ([]*domain.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]*domain.Game, 0, len(r.games))
	for _, g := range r.games {
		res = append(res, g)
	}
	return res, nil
}
