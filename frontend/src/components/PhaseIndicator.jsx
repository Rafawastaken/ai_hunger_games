import './PhaseIndicator.css';

const PHASES = [
    { id: 'waiting', label: 'A Aguardar', icon: '⏳' },
    { id: 'answers', label: 'Respostas', icon: '💬' },
    { id: 'debate', label: 'Debate', icon: '⚔️' },
    { id: 'voting', label: 'Votação', icon: '🗳️' },
    { id: 'results', label: 'Resultados', icon: '🏆' },
];

export function PhaseIndicator({ currentPhase = 'waiting' }) {
    const currentIndex = PHASES.findIndex(p => p.id === currentPhase);

    return (
        <div className="phase-indicator">
            {PHASES.map((phase, index) => {
                const isCompleted = index < currentIndex;
                const isCurrent = index === currentIndex;

                return (
                    <div
                        key={phase.id}
                        className={`phase-step ${isCompleted ? 'completed' : ''} ${isCurrent ? 'current' : ''}`}
                    >
                        <div className="phase-icon">
                            {isCompleted ? '✓' : phase.icon}
                        </div>
                        <span className="phase-label">{phase.label}</span>
                        {index < PHASES.length - 1 && <div className="phase-connector" />}
                    </div>
                );
            })}
        </div>
    );
}
