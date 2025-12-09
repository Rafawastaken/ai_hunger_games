package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/rafawastaken/ai-hunger-games/internal/domain"
	"github.com/rafawastaken/ai-hunger-games/internal/repository"
	"github.com/rafawastaken/ai-hunger-games/internal/service"
	"github.com/rafawastaken/ai-hunger-games/internal/usecase"
)

type GameHandler struct {
	gameRepo     repository.GameRepository
	createGameUC *usecase.CreateGameUseCase
	playRoundUC  *usecase.PlayRoundUseCase

	groqSvc service.GroqService // para streaming
}

func NewGameHandler(
	gameRepo repository.GameRepository,
	createGameUC *usecase.CreateGameUseCase,
	playRoundUC *usecase.PlayRoundUseCase,
	groqSvc service.GroqService,
) *GameHandler {
	return &GameHandler{
		gameRepo:     gameRepo,
		createGameUC: createGameUC,
		playRoundUC:  playRoundUC,
		groqSvc:      groqSvc,
	}
}

func (h *GameHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/games", h.handleGames)
	mux.HandleFunc("/games/", h.handleGameByID)
}

func (h *GameHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /games  -> cria jogo
// GET  /games  -> lista jogos
func (h *GameHandler) handleGames(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req struct {
			NumAgents  int `json:"num_agents"`
			MaxStrikes int `json:"max_strikes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		out, err := h.createGameUC.Execute(usecase.CreateGameInput{
			NumAgents:  req.NumAgents,
			MaxStrikes: req.MaxStrikes,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, out.Game)

	case http.MethodGet:
		games, err := h.gameRepo.List()
		if err != nil {
			http.Error(w, "error listing games", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, games)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET  /games/{id}                 -> estado do jogo
// POST /games/{id}/rounds          -> corre 1 ronda (sem streaming)
// POST /games/{id}/rounds/stream   -> corre 1 ronda em SSE
func (h *GameHandler) handleGameByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/games/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	gameID := parts[0]

	// /games/{id}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			game, err := h.gameRepo.Get(gameID)
			if err != nil {
				http.Error(w, "game not found", http.StatusNotFound)
				return
			}
			writeJSON(w, http.StatusOK, game)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// /games/{id}/rounds...
	if parts[1] == "rounds" {
		// /games/{id}/rounds/stream
		if len(parts) >= 3 && parts[2] == "stream" {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.handlePlayRoundStream(w, r, gameID)
			return
		}

		// /games/{id}/rounds (normal, sem streaming)
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Question string `json:"question"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
		defer cancel()

		out, err := h.playRoundUC.Execute(ctx, usecase.PlayRoundInput{
			GameID:   gameID,
			Question: req.Question,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		writeJSON(w, http.StatusOK, out)
		return
	}

	http.NotFound(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// === Helpers para SSE ===

func sseWriteEvent(w http.ResponseWriter, flusher http.Flusher, event string, payload any) error {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var data []byte
	var err error

	switch v := payload.(type) {
	case string:
		data = []byte(v)
	default:
		data, err = json.Marshal(v)
		if err != nil {
			return err
		}
	}

	if event != "" {
		if _, err := w.Write([]byte("event: " + event + "\n")); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte("data: ")); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if _, err := w.Write([]byte("\n\n")); err != nil {
		return err
	}

	flusher.Flush()
	return nil
}

// === Endpoint: /games/{id}/rounds/stream ===

func (h *GameHandler) handlePlayRoundStream(w http.ResponseWriter, r *http.Request, gameID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Ler pergunta do body
	var req struct {
		Question string `json:"question"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Question) == "" {
		http.Error(w, "question is required", http.StatusBadRequest)
		return
	}

	// Buscar jogo
	game, err := h.gameRepo.Get(gameID)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}
	if game.Status == domain.GameStatusFinished {
		http.Error(w, "game already finished", http.StatusBadRequest)
		return
	}

	activeAgents := game.ActiveAgents()
	if len(activeAgents) == 0 {
		http.Error(w, "no active agents in game", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 180*time.Second)
	defer cancel()

	round := &domain.Round{
		Index:    game.NextRoundIndex(),
		Question: req.Question,
	}

	// 1) Respostas iniciais (stream: event "answer")
	for _, agent := range activeAgents {
		select {
		case <-ctx.Done():
			return
		default:
		}

		text, err := h.groqSvc.GenerateAnswer(ctx, game, agent, req.Question)
		if err != nil {
			_ = sseWriteEvent(w, flusher, "error", map[string]string{"error": err.Error()})
			return
		}

		ans := domain.Answer{
			AgentID: agent.ID,
			Text:    text,
		}
		round.Answers = append(round.Answers, ans)

		_ = sseWriteEvent(w, flusher, "answer", ans)
	}

	_ = sseWriteEvent(w, flusher, "phase", map[string]string{"phase": "answers_done"})

	// 2) Debate (stream: event "debate")
	const debateTurns = 2 // reduzido para economizar tokens
	for turn := 1; turn <= debateTurns; turn++ {
		for _, agent := range activeAgents {
			select {
			case <-ctx.Done():
				return
			default:
			}

			msg, err := h.groqSvc.GenerateDebateMessage(ctx, game, round, agent)
			if err != nil {
				_ = sseWriteEvent(w, flusher, "error", map[string]string{"error": err.Error()})
				return
			}

			dm := domain.DebateMessage{
				AgentID: agent.ID,
				Turn:    turn,
				Text:    msg,
			}
			round.Debate = append(round.Debate, dm)

			_ = sseWriteEvent(w, flusher, "debate", dm)
		}
	}

	_ = sseWriteEvent(w, flusher, "phase", map[string]string{"phase": "debate_done"})

	// 3) Votos (stream: event "vote")
	votesCount := make(map[string]int)
	for _, agent := range activeAgents {
		votesCount[agent.ID] = 0
	}

	for _, agent := range activeAgents {
		select {
		case <-ctx.Done():
			return
		default:
		}

		targetID, justification, err := h.groqSvc.GenerateVote(ctx, game, round, agent)
		if err != nil {
			_ = sseWriteEvent(w, flusher, "error", map[string]string{"error": err.Error()})
			return
		}

		v := domain.Vote{
			VoterID:       agent.ID,
			TargetID:      targetID,
			Justification: justification,
		}
		round.Votes = append(round.Votes, v)
		if _, ok := votesCount[targetID]; ok {
			votesCount[targetID]++
		}

		_ = sseWriteEvent(w, flusher, "vote", v)
	}

	// 4) Determinar strikes (igual ao usecase normal)
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

	// 5) Atualizar jogo e mandar round_end
	game.Rounds = append(game.Rounds, round)
	if len(game.ActiveAgents()) <= 1 {
		game.Status = domain.GameStatusFinished
	} else {
		game.Status = domain.GameStatusRunning
	}

	if err := h.gameRepo.Update(game); err != nil {
		_ = sseWriteEvent(w, flusher, "error", map[string]string{"error": err.Error()})
		return
	}

	type roundEndPayload struct {
		Game  *domain.Game  `json:"game"`
		Round *domain.Round `json:"round"`
	}

	_ = sseWriteEvent(w, flusher, "round_end", roundEndPayload{
		Game:  game,
		Round: round,
	})
}
