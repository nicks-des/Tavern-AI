package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/tavern-ai/backend/internal/models"
	"github.com/tavern-ai/backend/internal/repository"
)

type SessionHandler struct {
	sessionRepo *repository.SessionRepo
	messageRepo *repository.MessageRepo
}

func NewSessionHandler(sr *repository.SessionRepo, mr *repository.MessageRepo) *SessionHandler {
	return &SessionHandler{sessionRepo: sr, messageRepo: mr}
}

func (h *SessionHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/sessions", h.listSessions)
	mux.HandleFunc("POST /api/sessions", h.createSession)
	mux.HandleFunc("GET /api/sessions/{id}", h.getSession)
	mux.HandleFunc("DELETE /api/sessions/{id}", h.deleteSession)

	mux.HandleFunc("POST /api/sessions/{id}/messages", h.sendMessage)
	mux.HandleFunc("GET /api/sessions/{id}/messages", h.listMessages)
}

func (h *SessionHandler) listSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.sessionRepo.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sessions == nil {
		sessions = []models.Session{}
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (h *SessionHandler) createSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID          string  `json:"id"`
		CharacterID string  `json:"characterId"`
		RoomID      *string `json:"roomId"`
		Title       string  `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.CharacterID == "" {
		writeError(w, http.StatusBadRequest, "characterId is required")
		return
	}
	if req.Title == "" {
		req.Title = "新对话"
	}

	sessionID := req.ID
	if sessionID == "" {
		sessionID = generateID()
	}

	now := time.Now()
	s := &models.Session{
		ID:          sessionID,
		CharacterID: req.CharacterID,
		RoomID:      req.RoomID,
		Title:       req.Title,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := h.sessionRepo.Create(s); err != nil {
		writeError(w, http.StatusInternalServerError, "create session: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *SessionHandler) getSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	session, err := h.sessionRepo.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	messages, err := h.messageRepo.ListBySession(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if messages == nil {
		messages = []models.Message{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session":  session,
		"messages": messages,
	})
}

func (h *SessionHandler) deleteSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.sessionRepo.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *SessionHandler) sendMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	var req struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Role == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "role and content are required")
		return
	}

	now := time.Now()
	msg := &models.Message{
		ID:        generateID(),
		SessionID: sessionID,
		Role:      req.Role,
		Content:   req.Content,
		Timestamp: now,
	}
	if err := h.messageRepo.Create(msg); err != nil {
		writeError(w, http.StatusInternalServerError, "save message: "+err.Error())
		return
	}

	h.sessionRepo.Touch(sessionID)
	writeJSON(w, http.StatusCreated, msg)
}

func (h *SessionHandler) listMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	messages, err := h.messageRepo.ListBySession(sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if messages == nil {
		messages = []models.Message{}
	}
	writeJSON(w, http.StatusOK, messages)
}
