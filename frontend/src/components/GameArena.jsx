import { useRef, useEffect } from 'react';
import { AgentCard } from './AgentCard';
import { MessageBubble } from './MessageBubble';
import { PhaseIndicator } from './PhaseIndicator';
import './GameArena.css';

export function GameArena({
    game,
    messages = [],
    currentPhase = 'waiting',
    speakingAgent = null,
    question = ''
}) {
    const messagesEndRef = useRef(null);

    // Auto-scroll to bottom when new messages arrive
    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [messages]);

    if (!game) return null;

    const agents = game.agents || [];
    const leftAgents = agents.filter((_, i) => i % 2 === 0);
    const rightAgents = agents.filter((_, i) => i % 2 === 1);

    return (
        <div className="game-arena">
            {/* Phase Indicator */}
            <div className="arena-header">
                <PhaseIndicator currentPhase={currentPhase} />
            </div>

            {/* Question Display */}
            {question && (
                <div className="arena-question">
                    <span className="question-label">Pergunta:</span>
                    <p className="question-text">{question}</p>
                </div>
            )}

            {/* Main Arena Layout */}
            <div className="arena-content">
                {/* Left Agents */}
                <div className="agents-column left">
                    {leftAgents.map(agent => (
                        <AgentCard
                            key={agent.id}
                            agent={agent}
                            isActive={!agent.eliminated}
                            isSpeaking={speakingAgent === agent.id}
                        />
                    ))}
                </div>

                {/* Center - Messages */}
                <div className="messages-column">
                    <div className="messages-container">
                        {messages.length === 0 ? (
                            <div className="messages-empty">
                                <span className="empty-icon">🔥</span>
                                <p>A aguardar início do debate...</p>
                            </div>
                        ) : (
                            messages.map((msg, index) => (
                                <MessageBubble
                                    key={`${msg.agentId}-${msg.phase}-${index}`}
                                    agentId={msg.agentId}
                                    agentName={msg.agentName}
                                    text={msg.text}
                                    phase={msg.phase}
                                    turn={msg.turn}
                                    voteTarget={msg.voteTarget}
                                    justification={msg.justification}
                                    isStreaming={index === messages.length - 1 && currentPhase !== 'results'}
                                />
                            ))
                        )}
                        <div ref={messagesEndRef} />
                    </div>
                </div>

                {/* Right Agents */}
                <div className="agents-column right">
                    {rightAgents.map(agent => (
                        <AgentCard
                            key={agent.id}
                            agent={agent}
                            isActive={!agent.eliminated}
                            isSpeaking={speakingAgent === agent.id}
                        />
                    ))}
                </div>
            </div>

            {/* Round Stats */}
            {game.rounds && game.rounds.length > 0 && (
                <div className="arena-stats">
                    <span className="stat">
                        <span className="stat-label">Ronda</span>
                        <span className="stat-value">{game.rounds.length}</span>
                    </span>
                    <span className="stat">
                        <span className="stat-label">Agentes Ativos</span>
                        <span className="stat-value">{agents.filter(a => !a.eliminated).length}</span>
                    </span>
                    <span className="stat">
                        <span className="stat-label">Eliminados</span>
                        <span className="stat-value text-red">{agents.filter(a => a.eliminated).length}</span>
                    </span>
                </div>
            )}
        </div>
    );
}
