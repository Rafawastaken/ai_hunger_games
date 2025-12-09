package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/rafawastaken/ai-hunger-games/internal/handler"
	"github.com/rafawastaken/ai-hunger-games/internal/repository"
	"github.com/rafawastaken/ai-hunger-games/internal/service"
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
	gameHandler.RegisterRoutes(mux)

	addr := ":8080"
	log.Printf("🔥 AI Hunger Games API a correr em http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("erro no servidor: %v", err)
	}
}
