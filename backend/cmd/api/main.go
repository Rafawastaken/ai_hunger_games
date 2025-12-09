package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/rafawastaken/ai-hunger-games/internal/handler"
	"github.com/rafawastaken/ai-hunger-games/internal/repository"
	"github.com/rafawastaken/ai-hunger-games/internal/service"
	"github.com/rafawastaken/ai-hunger-games/internal/static"
	"github.com/rafawastaken/ai-hunger-games/internal/usecase"
)

func main() {
	// Carregar .env (se existir)
	_ = godotenv.Load()

	apiKey := os.Getenv("GROQ_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GROQ_API_KEY")
	}
	if apiKey == "" {
		log.Fatal("GROQ_KEY ou GROQ_API_KEY não encontrados no ambiente/.env")
	}

	// Wiring de dependências
	gameRepo := repository.NewInMemoryGameRepository()
	groqSvc := service.NewGroqService(apiKey, "")

	createGameUC := usecase.NewCreateGameUseCase(gameRepo)
	playRoundUC := usecase.NewPlayRoundUseCase(gameRepo, groqSvc)

	gameHandler := handler.NewGameHandler(gameRepo, createGameUC, playRoundUC, groqSvc)

	mux := http.NewServeMux()

	// Registar rotas da API
	gameHandler.RegisterRoutes(mux)

	// Servir ficheiros estáticos do frontend embebido
	staticFS := static.GetFileSystem()
	fileServer := http.FileServer(staticFS)

	// Handler para SPA - serve index.html para rotas que não são API nem ficheiros estáticos
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Se é um pedido de API, deixar passar (já foi tratado pelos handlers)
		if strings.HasPrefix(r.URL.Path, "/health") ||
			strings.HasPrefix(r.URL.Path, "/games") {
			http.NotFound(w, r)
			return
		}

		// Tentar servir ficheiro estático
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Verificar se o ficheiro existe
		staticFSys := static.GetFS()
		cleanPath := strings.TrimPrefix(path, "/")
		if _, err := fs.Stat(staticFSys, cleanPath); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback para index.html (SPA routing)
		r.URL.Path = "/index.html"
		fileServer.ServeHTTP(w, r)
	})

	addr := ":8080"
	log.Printf("🔥 AI Hunger Games a correr em http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("erro no servidor: %v", err)
	}
}
