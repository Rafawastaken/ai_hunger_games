import { useState, useCallback } from 'react';
import { GameArena } from './components/GameArena';
import { createGame, playRoundStream } from './services/api';
import './index.css';
import './App.css';

function App() {
  // Game state
  const [game, setGame] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);

  // Round state
  const [messages, setMessages] = useState([]);
  const [currentPhase, setCurrentPhase] = useState('waiting');
  const [speakingAgent, setSpeakingAgent] = useState(null);
  const [question, setQuestion] = useState('');
  const [questionInput, setQuestionInput] = useState('');
  const [isRoundRunning, setIsRoundRunning] = useState(false);

  // Lobby settings
  const [numAgents, setNumAgents] = useState(4);
  const [maxStrikes, setMaxStrikes] = useState(2);

  // Get agent name from game
  const getAgentName = useCallback((agentId) => {
    if (!game) return agentId;
    const agent = game.agents.find(a => a.id === agentId);
    return agent?.name || agentId;
  }, [game]);

  // Create new game
  const handleCreateGame = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const newGame = await createGame(numAgents, maxStrikes);
      setGame(newGame);
      setMessages([]);
      setCurrentPhase('waiting');
      setQuestion('');
    } catch (err) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Play a round
  const handlePlayRound = async () => {
    if (!questionInput.trim() || !game || isRoundRunning) return;

    setIsRoundRunning(true);
    setError(null);
    setMessages([]);
    setQuestion(questionInput);
    setCurrentPhase('answers');

    const cleanup = playRoundStream(game.id, questionInput, {
      onAnswer: (answer) => {
        setSpeakingAgent(answer.agent_id);
        setMessages(prev => [...prev, {
          agentId: answer.agent_id,
          agentName: getAgentName(answer.agent_id),
          text: answer.text,
          phase: 'answer'
        }]);
      },

      onDebate: (debate) => {
        setCurrentPhase('debate');
        setSpeakingAgent(debate.agent_id);
        setMessages(prev => [...prev, {
          agentId: debate.agent_id,
          agentName: getAgentName(debate.agent_id),
          text: debate.text,
          phase: 'debate',
          turn: debate.turn
        }]);
      },

      onVote: (vote) => {
        setCurrentPhase('voting');
        setSpeakingAgent(vote.voter_id);
        setMessages(prev => [...prev, {
          agentId: vote.voter_id,
          agentName: getAgentName(vote.voter_id),
          phase: 'vote',
          voteTarget: getAgentName(vote.target_id),
          justification: vote.justification
        }]);
      },

      onPhase: (phase) => {
        if (phase === 'answers_done') {
          setCurrentPhase('debate');
        } else if (phase === 'debate_done') {
          setCurrentPhase('voting');
        }
      },

      onRoundEnd: (result) => {
        setCurrentPhase('results');
        setSpeakingAgent(null);
        setIsRoundRunning(false);
        setGame(result.game);
        setQuestionInput('');
      },

      onError: (err) => {
        setError(err);
        setIsRoundRunning(false);
        setSpeakingAgent(null);
      }
    });

    // Store cleanup for potential abort
    return cleanup;
  };

  // Check if game is finished
  const isGameFinished = game?.status === 'finished';
  const winner = isGameFinished ? game?.agents.find(a => !a.eliminated) : null;

  return (
    <div className="app">
      {/* Header */}
      <header className="app-header">
        <h1>AI Hunger Games</h1>
        <p className="tagline">Que os melhores argumentos vençam</p>
      </header>

      {/* Error Display */}
      {error && (
        <div className="error-banner">
          <span>⚠️ {error}</span>
          <button onClick={() => setError(null)}>✕</button>
        </div>
      )}

      {/* Main Content */}
      <main className="app-main">
        {!game ? (
          /* Lobby - Create Game */
          <div className="lobby">
            <div className="lobby-card card">
              <h2>Inicia um Novo Jogo</h2>
              <p className="lobby-description">
                Configura a arena e lança os agentes IA num debate épico onde apenas o mais convincente sobrevive.
              </p>

              <div className="lobby-settings">
                <div className="setting-group">
                  <label htmlFor="numAgents">Número de Agentes</label>
                  <div className="setting-input">
                    <button
                      className="btn-adjust"
                      onClick={() => setNumAgents(Math.max(2, numAgents - 1))}
                      disabled={numAgents <= 2}
                    >
                      −
                    </button>
                    <input
                      id="numAgents"
                      type="number"
                      value={numAgents}
                      onChange={(e) => setNumAgents(Math.max(2, Math.min(8, parseInt(e.target.value) || 2)))}
                      min="2"
                      max="8"
                    />
                    <button
                      className="btn-adjust"
                      onClick={() => setNumAgents(Math.min(8, numAgents + 1))}
                      disabled={numAgents >= 8}
                    >
                      +
                    </button>
                  </div>
                </div>

                <div className="setting-group">
                  <label htmlFor="maxStrikes">Strikes para Eliminação</label>
                  <div className="setting-input">
                    <button
                      className="btn-adjust"
                      onClick={() => setMaxStrikes(Math.max(1, maxStrikes - 1))}
                      disabled={maxStrikes <= 1}
                    >
                      −
                    </button>
                    <input
                      id="maxStrikes"
                      type="number"
                      value={maxStrikes}
                      onChange={(e) => setMaxStrikes(Math.max(1, Math.min(5, parseInt(e.target.value) || 1)))}
                      min="1"
                      max="5"
                    />
                    <button
                      className="btn-adjust"
                      onClick={() => setMaxStrikes(Math.min(5, maxStrikes + 1))}
                      disabled={maxStrikes >= 5}
                    >
                      +
                    </button>
                  </div>
                </div>
              </div>

              <button
                className="btn btn-primary start-btn"
                onClick={handleCreateGame}
                disabled={isLoading}
              >
                {isLoading ? 'A Criar Arena...' : '🔥 Iniciar Jogos'}
              </button>
            </div>

            <div className="lobby-rules card">
              <h3>Regras da Arena</h3>
              <ul>
                <li>🎯 Faz uma pergunta a todos os agentes</li>
                <li>💬 Os agentes respondem e debatem entre si</li>
                <li>🗳️ No final, votam na melhor resposta</li>
                <li>🔥 O menos votado recebe um <strong>strike</strong></li>
                <li>💀 Com {maxStrikes} strikes, o agente é <strong>eliminado</strong></li>
                <li>🏆 O último sobrevivente vence!</li>
              </ul>
            </div>
          </div>
        ) : isGameFinished && winner ? (
          /* Winner Screen */
          <div className="winner-screen">
            <div className="winner-card card card-glow">
              <div className="winner-crown">👑</div>
              <h2>Vencedor!</h2>
              <div className="winner-name">{winner.name}</div>
              <p className="winner-message">
                Sobreviveu a {game.rounds?.length || 0} rondas de debate intenso!
              </p>
              <button
                className="btn btn-primary"
                onClick={() => {
                  setGame(null);
                  setMessages([]);
                  setCurrentPhase('waiting');
                }}
              >
                🔄 Novo Jogo
              </button>
            </div>
          </div>
        ) : (
          /* Game Arena */
          <div className="arena-wrapper">
            <GameArena
              game={game}
              messages={messages}
              currentPhase={currentPhase}
              speakingAgent={speakingAgent}
              question={question}
            />

            {/* Question Input */}
            <div className="question-input-wrapper">
              <div className="question-input-container">
                <input
                  type="text"
                  value={questionInput}
                  onChange={(e) => setQuestionInput(e.target.value)}
                  placeholder="Escreve uma pergunta para os agentes debaterem..."
                  disabled={isRoundRunning}
                  onKeyDown={(e) => e.key === 'Enter' && handlePlayRound()}
                />
                <button
                  className="btn btn-primary"
                  onClick={handlePlayRound}
                  disabled={!questionInput.trim() || isRoundRunning}
                >
                  {isRoundRunning ? '⏳ Em Debate...' : '⚔️ Lançar Pergunta'}
                </button>
              </div>

              {/* Quick Questions */}
              {!isRoundRunning && (
                <div className="quick-questions">
                  <span className="quick-label">Sugestões:</span>
                  {[
                    'Qual é o sentido da vida?',
                    'A IA vai substituir os humanos?',
                    'Qual é a melhor linguagem de programação?',
                  ].map((q, i) => (
                    <button
                      key={i}
                      className="quick-btn"
                      onClick={() => setQuestionInput(q)}
                    >
                      {q}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="app-footer">
        <p>Criado com 🔥 por <a href="https://github.com/rafawastaken" target="_blank" rel="noopener noreferrer">rafawastaken</a></p>
      </footer>
    </div>
  );
}

export default App;
