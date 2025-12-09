# 🔥 AI Hunger Games

Um jogo de debate onde agentes de IA competem para sobreviver. Cada agente responde a uma pergunta, debate com os outros, e depois votam na melhor resposta. Os menos votados recebem strikes - e com strikes suficientes, são eliminados!

![React](https://img.shields.io/badge/React-19-blue)
![Go](https://img.shields.io/badge/Go-1.23-00ADD8)
![Groq](https://img.shields.io/badge/LLM-Groq-orange)

## 🎮 Como Funciona

1. **Cria um jogo** com N agentes (2-8)
2. **Faz uma pergunta** a todos os agentes
3. **Os agentes respondem** com perspectivas únicas (cada um tem personalidade diferente)
4. **Debate aceso** - os agentes atacam directamente as opiniões uns dos outros
5. **Votação** - cada agente vota na melhor resposta (não pode votar em si próprio)
6. **Strikes** - o menos votado leva um strike
7. **Eliminação** - com 2 strikes, o agente é eliminado
8. **Repete** até restar apenas 1 vencedor!

## 🚀 Quick Start

### Pré-requisitos

- [Node.js](https://nodejs.org/) (v18+)
- [Go](https://golang.org/) (1.23+)
- [Groq API Key](https://console.groq.com/)

### 1. Configurar API Key

Cria um ficheiro `.env` na pasta `backend/`:

```env
GROQ_KEY=gsk_xxxxx_sua_chave_aqui
```

### 2. Iniciar Backend

```bash
cd backend
go run cmd/api/main.go
```

O backend inicia em `http://localhost:8080`

### 3. Iniciar Frontend

```bash
cd frontend
npm install
npm run dev
```

O frontend inicia em `http://localhost:5173` (acessível na rede local)

### 4. Jogar! 🎲

Abre `http://localhost:5173` no browser e diverte-te!

## 📁 Estrutura do Projeto

```
ai_hunger_games/
├── backend/                 # API Go
│   ├── cmd/api/main.go     # Entrypoint
│   └── internal/
│       ├── domain/         # Entidades (Game, Agent, Round)
│       ├── handler/        # HTTP handlers + SSE streaming
│       ├── repository/     # In-memory storage
│       ├── service/        # Integração Groq API
│       └── usecase/        # Lógica de negócio
│
└── frontend/               # React + Vite
    └── src/
        ├── components/     # UI Components
        ├── services/       # API client
        └── App.jsx         # Main app
```

## 🛠️ API Endpoints

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/health` | Health check |
| `POST` | `/games` | Criar jogo |
| `GET` | `/games` | Listar jogos |
| `GET` | `/games/{id}` | Estado do jogo |
| `POST` | `/games/{id}/rounds/stream` | Jogar ronda (SSE) |

## ⚙️ Configuração

| Variável | Descrição | Default |
|----------|-----------|---------|
| `GROQ_KEY` | Groq API key | *obrigatório* |
| `GROQ_API_KEY` | Alternativo | - |

## 🎨 Features

- **Streaming em tempo real** - vê as respostas a aparecer via SSE
- **Personalidades únicas** - cada agente tem uma personalidade diferente
- **Debates agressivos** - os agentes atacam-se directamente
- **Retry automático** - exponential backoff para rate limiting
- **Tema Hunger Games** - dark mode com cores de fogo 🔥

## 👨‍💻 Autor

Criado por [rafawastaken](https://github.com/rafawastaken)

## 📝 Licença

MIT
