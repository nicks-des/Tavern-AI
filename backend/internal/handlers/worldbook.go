package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/tavern-ai/backend/internal/models"
	"github.com/tavern-ai/backend/internal/repository"
)

type WorldBookHandler struct {
	repo *repository.WorldBookRepo
}

func NewWorldBookHandler(repo *repository.WorldBookRepo) *WorldBookHandler {
	return &WorldBookHandler{repo: repo}
}

func (h *WorldBookHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/characters/{id}/worldbook", h.listEntries)
	mux.HandleFunc("POST /api/characters/{id}/worldbook", h.createEntry)
	mux.HandleFunc("PUT /api/worldbook/{id}", h.updateEntry)
	mux.HandleFunc("DELETE /api/worldbook/{id}", h.deleteEntry)
}

func (h *WorldBookHandler) listEntries(w http.ResponseWriter, r *http.Request) {
	charID := r.PathValue("id")
	entries, err := h.repo.ListByCharacter(charID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if entries == nil {
		entries = []models.WorldBookEntry{}
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *WorldBookHandler) createEntry(w http.ResponseWriter, r *http.Request) {
	charID := r.PathValue("id")

	var req struct {
		Keywords string `json:"keywords"`
		Content  string `json:"content"`
		Enabled  bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Keywords == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "keywords and content are required")
		return
	}

	entry := &models.WorldBookEntry{
		ID:          generateID(),
		CharacterID: charID,
		Keywords:    req.Keywords,
		Content:     req.Content,
		Enabled:     true,
		CreatedAt:   time.Now(),
	}
	if err := h.repo.Create(entry); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}

func (h *WorldBookHandler) updateEntry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		Keywords string `json:"keywords"`
		Content  string `json:"content"`
		Enabled  *bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	entry, err := h.repo.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}

	if req.Keywords != "" {
		entry.Keywords = req.Keywords
	}
	if req.Content != "" {
		entry.Content = req.Content
	}
	if req.Enabled != nil {
		entry.Enabled = *req.Enabled
	}

	if err := h.repo.Update(entry); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

func (h *WorldBookHandler) deleteEntry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.repo.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
