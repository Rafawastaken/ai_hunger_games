import './MessageBubble.css';

const PHASE_STYLES = {
    answer: {
        label: 'Resposta',
        className: 'answer'
    },
    debate: {
        label: 'Debate',
        className: 'debate'
    },
    vote: {
        label: 'Voto',
        className: 'vote'
    }
};

const AGENT_COLORS = [
    '#ff6b35', '#3b82f6', '#10b981', '#f59e0b',
    '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16',
];

export function MessageBubble({
    agentId,
    agentName,
    text,
    phase = 'answer',
    turn,
    isStreaming = false,
    voteTarget,
    justification
}) {
    const agentNumber = parseInt(agentId?.replace('agent-', ''), 10) || 1;
    const color = AGENT_COLORS[(agentNumber - 1) % AGENT_COLORS.length];
    const phaseStyle = PHASE_STYLES[phase] || PHASE_STYLES.answer;

    return (
        <div
            className={`message-bubble ${phaseStyle.className} ${isStreaming ? 'streaming' : ''}`}
            style={{ '--agent-color': color }}
        >
            <div className="message-header">
                <div className="message-agent">
                    <span className="agent-dot" />
                    <span className="agent-label">{agentName || agentId}</span>
                </div>
                <div className="message-meta">
                    {turn && <span className="turn-badge">Turno {turn}</span>}
                    <span className="phase-badge">{phaseStyle.label}</span>
                </div>
            </div>

            <div className="message-content">
                {phase === 'vote' ? (
                    <div className="vote-content">
                        <div className="vote-target">
                            <span className="vote-label">Votou em:</span>
                            <span className="vote-value">{voteTarget}</span>
                        </div>
                        {justification && (
                            <p className="vote-justification">{justification}</p>
                        )}
                    </div>
                ) : (
                    <p>{text}</p>
                )}
                {isStreaming && <span className="typing-cursor" />}
            </div>
        </div>
    );
}
