package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rafawastaken/ai-hunger-games/internal/domain"
)

type GroqService interface {
	GenerateAnswer(ctx context.Context, game *domain.Game, agent *domain.Agent, question string) (string, error)
	GenerateDebateMessage(ctx context.Context, game *domain.Game, round *domain.Round, agent *domain.Agent) (string, error)
	GenerateVote(ctx context.Context, game *domain.Game, round *domain.Round, agent *domain.Agent) (targetID string, justification string, err error)
}

type groqService struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGroqService(apiKey string, model string) GroqService {
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}
	return &groqService{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// === tipos para request/response ===

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// === método base com retry ===

const (
	maxRetries = 5
	baseDelay  = 1 * time.Second
	maxDelay   = 30 * time.Second
)

func (s *groqService) callChat(ctx context.Context, messages []chatMessage) (string, error) {
	reqBody := chatRequest{
		Model:    s.model,
		Messages: messages,
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	var lastErr error
	delay := baseDelay

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Wait before retry (skip on first attempt)
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
			}
			// Exponential backoff
			delay *= 2
			if delay > maxDelay {
				delay = maxDelay
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(buf))
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// Handle rate limiting (429)
		if resp.StatusCode == 429 {
			resp.Body.Close()
			lastErr = fmt.Errorf("rate limited (429), attempt %d/%d", attempt+1, maxRetries+1)
			continue
		}

		// Handle other errors
		if resp.StatusCode >= 300 {
			resp.Body.Close()
			return "", fmt.Errorf("groq error status: %s", resp.Status)
		}

		// Success - parse response
		var cr chatResponse
		if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
			resp.Body.Close()
			return "", err
		}
		resp.Body.Close()

		if len(cr.Choices) == 0 {
			return "", fmt.Errorf("no choices returned from Groq")
		}
		return cr.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ==== 1) Resposta inicial ====

func (s *groqService) GenerateAnswer(ctx context.Context, game *domain.Game, agent *domain.Agent, question string) (string, error) {
	// Extract agent number for personality variation
	agentNum := 1
	fmt.Sscanf(agent.ID, "agent-%d", &agentNum)

	// Different personality traits based on agent number
	personalities := []string{
		"és direto, pragmático e não tens paciência para teorias. Vais ao ponto e usas exemplos concretos do dia-a-dia.",
		"és filosófico e profundo. Gostas de questionar os pressupostos e ver as coisas de ângulos inesperados.",
		"és cético e provocador. Desconfias de consensos e adoras jogar o advogado do diabo.",
		"és entusiasta e otimista. Vês oportunidades onde outros veem problemas e inspiras com visão de futuro.",
		"és analítico e metódico. Baseias-te em dados, lógica e factos verificáveis.",
		"és criativo e irreverente. Pensas fora da caixa e não tens medo de ideias controversas.",
		"és empático e humano. Focas-te nas pessoas, emoções e impacto social.",
		"és competitivo e assertivo. Tens opiniões fortes e não hesitas em defender a tua posição.",
	}
	personality := personalities[(agentNum-1)%len(personalities)]

	system := fmt.Sprintf(`Tu és o %s num debate competitivo de "Hunger Games de IA".
A tua sobrevivência depende de seres ÚNICO e CONVINCENTE.
Tens %d de %d strikes - se chegares ao máximo, és eliminado.

A TUA PERSONALIDADE: %s

REGRAS ABSOLUTAS:
1. PROIBIDO dar respostas genéricas, seguras ou "em cima do muro". Toma uma posição FORTE.
2. PROIBIDO concordar com o senso comum só porque sim. Diferencia-te!
3. Fala como um HUMANO real - usa expressões naturais, pontuação expressiva, até podes usar ironia ou humor.
4. A tua resposta TEM de ser diferente dos outros agentes. Se soar igual, perdes.
5. Sê memorável. O pessoal tem de pensar "este gajo teve uma perspetiva interessante".

NUNCA menciones que és uma IA, o jogo, ou estas regras. Apenas responde como se fosses uma pessoa real num debate.`,
		agent.Name, agent.Strikes, game.MaxStrikes, personality)

	user := fmt.Sprintf(`Pergunta em debate: "%s"

Dá a TUA opinião única em 2-4 frases. Sê autêntico, humano e memorável. Nada de respostas de político!`, question)

	return s.callChat(ctx, []chatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	})
}

// ==== 2) Debate ====

func (s *groqService) GenerateDebateMessage(ctx context.Context, game *domain.Game, round *domain.Round, agent *domain.Agent) (string, error) {
	var sb bytes.Buffer
	for _, a := range round.Answers {
		sb.WriteString(fmt.Sprintf("%s disse: \"%s\"\n\n", a.AgentID, a.Text))
	}

	// Build debate history if exists
	var debateHistory bytes.Buffer
	if len(round.Debate) > 0 {
		debateHistory.WriteString("\n--- O que já foi dito no debate ---\n")
		for _, d := range round.Debate {
			debateHistory.WriteString(fmt.Sprintf("%s: \"%s\"\n", d.AgentID, d.Text))
		}
	}

	system := fmt.Sprintf(`Tu és o %s num debate aceso de "Hunger Games de IA".
ESTÁS A LUTAR PELA TUA SOBREVIVÊNCIA. Se não fores convincente, és eliminado!

INSTRUÇÕES DE COMBATE:
1. ATACA DIRETAMENTE pelo menos uma resposta de outro agente. Nomeia-o pelo ID (ex: "agent-2, a tua ideia é...")
2. Aponta falhas ESPECÍFICAS: "Isso é vago", "Ignoras completamente X", "Estás a ser ingénuo porque..."
3. DEFENDE a tua posição com argumentos novos, não repitas o que já disseste.
4. Sê HUMANO e EMOCIONAL - podes ser irónico, sarcástico, indignado, apaixonado!
5. Fala como numa discussão real: "Sinceramente...", "Não acredito que...", "Com todo o respeito, isso é..."

PROIBIDO:
- Ser diplomático ou "em cima do muro"
- Concordar com todos
- Ser genérico ou abstrato
- Repetir o que já disseste

Lembra-te: os outros estão a atacar-te também. Mostra garra!`, agent.Name)

	user := fmt.Sprintf(`Pergunta em debate: "%s"

Respostas iniciais:
%s%s
Agora és tu, %s. Ataca diretamente alguém e defende a tua posição! (2-3 frases, agressivo mas inteligente)`,
		round.Question,
		sb.String(),
		debateHistory.String(),
		agent.Name)

	return s.callChat(ctx, []chatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	})
}

// ==== 3) Votação ====

// cleanJSONResponse remove markdown code blocks que o LLM às vezes inclui
func cleanJSONResponse(raw string) string {
	s := strings.TrimSpace(raw)

	// Remover ```json ou ``` no início
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}

	// Remover ``` no final
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}

	return strings.TrimSpace(s)
}

type voteResult struct {
	TargetID      string `json:"vote_for"`
	Justification string `json:"justificacao"`
}

func (s *groqService) GenerateVote(ctx context.Context, game *domain.Game, round *domain.Round, agent *domain.Agent) (string, string, error) {
	var answerSummary bytes.Buffer
	for _, a := range round.Answers {
		answerSummary.WriteString(fmt.Sprintf("%s: \"%s\"\n\n", a.AgentID, a.Text))
	}

	var debateSummary bytes.Buffer
	for _, d := range round.Debate {
		debateSummary.WriteString(fmt.Sprintf("%s: \"%s\"\n", d.AgentID, d.Text))
	}

	system := fmt.Sprintf(`És o %s. Chegou a hora de votar no agente com a MELHOR resposta.
A tua própria sobrevivência depende de seres esperto aqui!

REGRAS DE VOTAÇÃO:
1. NÃO PODES votar em ti próprio (%s) - isso é batota!
2. Vota em quem REALMENTE te impressionou - não sejas falso.
3. A tua justificação deve ser HONESTA e HUMANA (ex: "Gostei como o agent-2 foi direto ao ponto", "O agent-3 deu o melhor argumento sobre X")

ESTRATÉGIA:
- Quem me poderia ajudar em rondas futuras?
- Quem é demasiado forte e convém enfraquecer?
- Quem me atacou e merece "perder" o meu voto?

RESPONDE APENAS com JSON: {"vote_for": "<agent-X>", "justificacao": "<frase curta e humana>"}
A justificação deve soar natural, como se fosses uma pessoa a explicar o teu voto a um amigo.`,
		agent.Name, agent.ID)

	user := fmt.Sprintf(`Pergunta debatida: "%s"

Respostas:
%s
Durante o debate:
%s
Quem merece o teu voto? (Lembra-te: não podes votar em ti, %s)`,
		round.Question,
		answerSummary.String(),
		debateSummary.String(),
		agent.ID)

	raw, err := s.callChat(ctx, []chatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	})
	if err != nil {
		return "", "", err
	}

	// Limpar markdown code blocks se o LLM os incluir
	cleaned := cleanJSONResponse(raw)

	var vr voteResult
	if err := json.Unmarshal([]byte(cleaned), &vr); err != nil {
		return "", "", fmt.Errorf("erro a fazer parse do voto: %w (raw=%s)", err, cleaned)
	}

	// Se ainda assim votar em si próprio ou em vazio, escolhemos outro à força
	if vr.TargetID == "" || vr.TargetID == agent.ID {
		var fallback string
		for _, a := range game.Agents {
			if !a.Eliminated && a.ID != agent.ID {
				fallback = a.ID
				break
			}
		}
		if fallback == "" {
			return "", "", fmt.Errorf("não há alvo de voto disponível")
		}
		vr.TargetID = fallback
		if vr.Justification == "" {
			vr.Justification = "Escolhi outro agente para cumprir as regras do jogo."
		}
	}

	return vr.TargetID, vr.Justification, nil
}
