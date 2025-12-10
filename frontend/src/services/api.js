// API calls go through Vite proxy in development
const API_BASE = '/api';

/**
 * Create a new game
 */
export async function createGame(numAgents = 4, maxStrikes = 2) {
    const response = await fetch(`${API_BASE}/games`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ num_agents: numAgents, max_strikes: maxStrikes })
    });

    if (!response.ok) {
        throw new Error(`Failed to create game: ${response.statusText}`);
    }

    return response.json();
}

/**
 * Get game state
 */
export async function getGame(gameId) {
    const response = await fetch(`${API_BASE}/games/${gameId}`);

    if (!response.ok) {
        throw new Error(`Failed to get game: ${response.statusText}`);
    }

    return response.json();
}

/**
 * List all games
 */
export async function listGames() {
    const response = await fetch(`${API_BASE}/games`);

    if (!response.ok) {
        throw new Error(`Failed to list games: ${response.statusText}`);
    }

    return response.json();
}

/**
 * Play a round with SSE streaming
 * @param {string} gameId 
 * @param {string} question 
 * @param {Object} callbacks - Event callbacks
 * @param {Function} callbacks.onAnswer - Called when an agent answers
 * @param {Function} callbacks.onDebate - Called when a debate message arrives
 * @param {Function} callbacks.onVote - Called when a vote is cast
 * @param {Function} callbacks.onPhase - Called when phase changes
 * @param {Function} callbacks.onRoundEnd - Called when round ends
 * @param {Function} callbacks.onError - Called on error
 * @returns {Function} Cleanup function to close the connection
 */
export function playRoundStream(gameId, question, callbacks) {
    const {
        onAnswer = () => { },
        onDebate = () => { },
        onVote = () => { },
        onJudgeVote = () => { },
        onPhase = () => { },
        onRoundEnd = () => { },
        onError = () => { }
    } = callbacks;

    // We need to use fetch with ReadableStream since EventSource doesn't support POST
    const controller = new AbortController();

    fetch(`${API_BASE}/games/${gameId}/rounds/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ question }),
        signal: controller.signal
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Stream failed: ${response.statusText}`);
            }

            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let buffer = '';

            function processBuffer() {
                const lines = buffer.split('\n');
                buffer = lines.pop() || '';

                let currentEvent = '';

                for (const line of lines) {
                    if (line.startsWith('event: ')) {
                        currentEvent = line.slice(7).trim();
                    } else if (line.startsWith('data: ')) {
                        const data = line.slice(6);
                        try {
                            const parsed = JSON.parse(data);

                            switch (currentEvent) {
                                case 'answer':
                                    onAnswer(parsed);
                                    break;
                                case 'debate':
                                    onDebate(parsed);
                                    break;
                                case 'vote':
                                    onVote(parsed);
                                    break;
                                case 'judge_vote':
                                    onJudgeVote(parsed);
                                    break;
                                case 'phase':
                                    onPhase(parsed.phase);
                                    break;
                                case 'round_end':
                                    onRoundEnd(parsed);
                                    break;
                                case 'error':
                                    onError(parsed.error || 'Unknown error');
                                    break;
                            }
                        } catch {
                            // Not JSON, ignore
                        }
                        currentEvent = '';
                    }
                }
            }

            function read() {
                reader.read().then(({ done, value }) => {
                    if (done) return;

                    buffer += decoder.decode(value, { stream: true });
                    processBuffer();
                    read();
                }).catch(err => {
                    if (err.name !== 'AbortError') {
                        onError(err.message);
                    }
                });
            }

            read();
        })
        .catch(err => {
            if (err.name !== 'AbortError') {
                onError(err.message);
            }
        });

    // Return cleanup function
    return () => controller.abort();
}
