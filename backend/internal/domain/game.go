package domain

type Agent struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Strikes    int    `json:"strikes"`
	Eliminated bool   `json:"eliminated"`
}

type Answer struct {
	AgentID string `json:"agent_id"`
	Text    string `json:"text"`
}

type DebateMessage struct {
	AgentID string `json:"agent_id"`
	Turn    int    `json:"turn"`
	Text    string `json:"text"`
}

type Vote struct {
	VoterID       string `json:"voter_id"`
	TargetID      string `json:"target_id"`
	Justification string `json:"justification"`
}

type Round struct {
	Index      int             `json:"index"`
	Question   string          `json:"question"`
	Answers    []Answer        `json:"answers"`
	Debate     []DebateMessage `json:"debate"`
	Votes      []Vote          `json:"votes"`
	Eliminated []string        `json:"eliminated"`
}

type GameStatus string

const (
	GameStatusWaiting  GameStatus = "waiting"
	GameStatusRunning  GameStatus = "running"
	GameStatusFinished GameStatus = "finished"
)

type Game struct {
	ID         string     `json:"id"`
	Agents     []*Agent   `json:"agents"`
	Rounds     []*Round   `json:"rounds"`
	MaxStrikes int        `json:"max_strikes"`
	Status     GameStatus `json:"status"`
}

// Helpers

func (g *Game) ActiveAgents() []*Agent {
	var res []*Agent
	for _, a := range g.Agents {
		if !a.Eliminated {
			res = append(res, a)
		}
	}
	return res
}

func (g *Game) NextRoundIndex() int {
	return len(g.Rounds) + 1
}
