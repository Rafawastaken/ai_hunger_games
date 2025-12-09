import './AgentCard.css';

const AGENT_COLORS = [
    '#ff6b35', // Fire orange
    '#3b82f6', // Blue
    '#10b981', // Emerald
    '#f59e0b', // Amber
    '#8b5cf6', // Purple
    '#ec4899', // Pink
    '#06b6d4', // Cyan
    '#84cc16', // Lime
];

export function AgentCard({ agent, isActive = false, isSpeaking = false }) {
    const agentNumber = parseInt(agent.id.replace('agent-', ''), 10) || 1;
    const color = AGENT_COLORS[(agentNumber - 1) % AGENT_COLORS.length];

    const strikes = Array.from({ length: agent.strikes }, (_, i) => i);
    const maxStrikesDisplay = 2; // Visual indicator

    return (
        <div
            className={`agent-card ${agent.eliminated ? 'eliminated' : ''} ${isActive ? 'active' : ''} ${isSpeaking ? 'speaking' : ''}`}
            style={{ '--agent-color': color }}
        >
            <div className="agent-avatar">
                <span className="agent-number">{agentNumber}</span>
                {isSpeaking && <div className="speaking-indicator" />}
            </div>

            <div className="agent-info">
                <span className="agent-name">{agent.name}</span>
                <div className="agent-strikes">
                    {[...Array(maxStrikesDisplay)].map((_, i) => (
                        <span
                            key={i}
                            className={`strike-icon ${i < agent.strikes ? 'active' : ''}`}
                            title={i < agent.strikes ? 'Strike!' : 'No strike'}
                        >
                            🔥
                        </span>
                    ))}
                </div>
            </div>

            {agent.eliminated && (
                <div className="eliminated-badge">
                    <span>ELIMINADO</span>
                </div>
            )}
        </div>
    );
}
