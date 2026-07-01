package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tavern-ai/backend/internal/config"
	"github.com/tavern-ai/backend/internal/database"
	"github.com/tavern-ai/backend/internal/handlers"
	"github.com/tavern-ai/backend/internal/llm"
	"github.com/tavern-ai/backend/internal/repository"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	fmt.Printf("Database initialized at: %s\n", cfg.DataDir)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	charRepo := repository.NewCharacterRepo(db)
	charHandler := handlers.NewCharacterHandler(charRepo)
	charHandler.Register(mux)

	sessionRepo := repository.NewSessionRepo(db)
	messageRepo := repository.NewMessageRepo(db)
	sessionHandler := handlers.NewSessionHandler(sessionRepo, messageRepo)
	sessionHandler.Register(mux)

	wbRepo := repository.NewWorldBookRepo(db)
	wbHandler := handlers.NewWorldBookHandler(wbRepo)
	wbHandler.Register(mux)

	var llmClient *llm.Client
	if cfg.OpenAIKey != "" {
		llmClient = llm.NewClient(llm.Config{
			APIKey:  cfg.OpenAIKey,
			BaseURL: cfg.OpenAIURL,
			Model:   cfg.ModelName,
		})
		fmt.Println("LLM: OpenAI client ready")
	} else {
		fmt.Println("LLM: no API key set, using mock mode")
	}

	chatHandler := handlers.NewChatHandler(sessionRepo, messageRepo, charRepo, wbRepo, llmClient)
	chatHandler.Register(mux)

	addr := ":" + cfg.HTTPPort
	fmt.Printf("Tavern AI backend starting on %s\n", addr)
	fmt.Println("Endpoints: /health, /api/characters, /api/sessions, /api/sessions/{id}/chat")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
