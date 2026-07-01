package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tavern-ai/backend/internal/llm"
	"github.com/tavern-ai/backend/internal/models"
	"github.com/tavern-ai/backend/internal/repository"
)

type ChatHandler struct {
	sessionRepo   *repository.SessionRepo
	messageRepo   *repository.MessageRepo
	characterRepo *repository.CharacterRepo
	llmClient     *llm.Client
}

func NewChatHandler(
	sr *repository.SessionRepo,
	mr *repository.MessageRepo,
	cr *repository.CharacterRepo,
	lc *llm.Client,
) *ChatHandler {
	return &ChatHandler{
		sessionRepo:   sr,
		messageRepo:   mr,
		characterRepo: cr,
		llmClient:     lc,
	}
}

func (h *ChatHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/sessions/{id}/chat", h.handleChat)
}

func (h *ChatHandler) handleChat(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	session, err := h.sessionRepo.GetByID(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	character, err := h.characterRepo.GetByID(session.CharacterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "character not found")
		return
	}

	history, err := h.messageRepo.ListBySession(sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load history failed")
		return
	}

	userMsg := &models.Message{
		ID:        generateID(),
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	}
	if err := h.messageRepo.Create(userMsg); err != nil {
		writeError(w, http.StatusInternalServerError, "save message failed")
		return
	}

	messages := h.llmClient.BuildMessages(character, history, req.Message)

	if h.llmClient == nil {
		h.handleMockChat(w, sessionID, character, req.Message)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	fullResponse, err := h.llmClient.ChatStream(messages, func(token string) {
		fmt.Fprintf(w, "data: %s\n\n", token)
		flusher.Flush()
	})

	if err != nil {
		fmt.Fprintf(w, "data: [ERROR] %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	aiMsg := &models.Message{
		ID:        generateID(),
		SessionID: sessionID,
		Role:      "assistant",
		Content:   fullResponse,
		Timestamp: time.Now(),
	}
	h.messageRepo.Create(aiMsg)
	h.sessionRepo.Touch(sessionID)

	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func (h *ChatHandler) handleMockChat(w http.ResponseWriter, sessionID string, character *models.Character, msg string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, _ := w.(http.Flusher)

	mockReply := fmt.Sprintf("这是对「%s」的模拟回复。我扮演的角色是 %s。配置 OPENAI_API_KEY 环境变量后即可启用真实 AI 对话。", msg, character.Name)

	for _, ch := range mockReply {
		fmt.Fprintf(w, "data: %s\n\n", string(ch))
		if flusher != nil {
			flusher.Flush()
		}
	}

	aiMsg := &models.Message{
		ID:        generateID(),
		SessionID: sessionID,
		Role:      "assistant",
		Content:   mockReply,
		Timestamp: time.Now(),
	}
	h.messageRepo.Create(aiMsg)
	h.sessionRepo.Touch(sessionID)

	fmt.Fprint(w, "data: [DONE]\n\n")
	if flusher != nil {
		flusher.Flush()
	}
}
