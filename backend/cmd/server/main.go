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

	addr := ":" + cfg.HTTPPort
	fmt.Printf("Tavern AI backend starting on %s\n", addr)

	handler := corsMiddleware(mux)

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

	roomRepo := repository.NewRoomRepo(db)
	roomHandler := handlers.NewRoomHandler(roomRepo, charRepo)
	roomHandler.Register(mux)

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

	chatHandler := handlers.NewChatHandler(sessionRepo, messageRepo, charRepo, wbRepo, roomRepo, llmClient)
	chatHandler.Register(mux)

	fmt.Println("Endpoints: /health, /api/characters, /api/sessions, /api/sessions/{id}/chat")
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			return
		}
		next.ServeHTTP(w, r)
	})
}
