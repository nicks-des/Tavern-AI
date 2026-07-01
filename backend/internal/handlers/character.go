package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/tavern-ai/backend/internal/models"
	"github.com/tavern-ai/backend/internal/repository"
)

type CharacterHandler struct {
	repo *repository.CharacterRepo
}

func NewCharacterHandler(repo *repository.CharacterRepo) *CharacterHandler {
	return &CharacterHandler{repo: repo}
}

func (h *CharacterHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/characters", h.listCharacters)
	mux.HandleFunc("POST /api/characters", h.createCharacter)
	mux.HandleFunc("GET /api/characters/{id}", h.getCharacter)
	mux.HandleFunc("PUT /api/characters/{id}", h.updateCharacter)
	mux.HandleFunc("DELETE /api/characters/{id}", h.deleteCharacter)
}

func (h *CharacterHandler) listCharacters(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")

	var chars []models.Character
	var err error
	if roomID != "" {
		chars, err = h.repo.ListByRoom(roomID)
	} else {
		chars, err = h.repo.ListGlobal()
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if chars == nil {
		chars = []models.Character{}
	}
	writeJSON(w, http.StatusOK, chars)
}

func (h *CharacterHandler) getCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, err := h.repo.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "character not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

type createCharacterRequest struct {
	Name            string              `json:"name"`
	Avatar          string              `json:"avatar"`
	Portrait        string              `json:"portrait"`
	Catchphrase     string              `json:"catchphrase"`
	Description     string              `json:"description"`
	Personality     string              `json:"personality"`
	Scenario        string              `json:"scenario"`
	FirstMessage    string              `json:"firstMessage"`
	ExampleDialogue string              `json:"exampleDialogue"`
	Tags            []string            `json:"tags"`
	Scope           models.CharacterScope `json:"scope"`
	RoomID          *string             `json:"roomId"`
}

func (h *CharacterHandler) createCharacter(w http.ResponseWriter, r *http.Request) {
	var req createCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}

	if req.Name == "" || req.Description == "" {
		writeError(w, http.StatusBadRequest, "name and description are required")
		return
	}
	if req.Scope == "" {
		req.Scope = models.ScopeGlobal
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}

	c := &models.Character{
		ID:              generateID(),
		Name:            req.Name,
		Avatar:          req.Avatar,
		Portrait:        req.Portrait,
		Catchphrase:     req.Catchphrase,
		Description:     req.Description,
		Personality:     req.Personality,
		Scenario:        req.Scenario,
		FirstMessage:    req.FirstMessage,
		ExampleDialogue: req.ExampleDialogue,
		Tags:            req.Tags,
		Scope:           req.Scope,
		RoomID:          req.RoomID,
		CreatedAt:       time.Now(),
	}

	if err := h.repo.Create(c); err != nil {
		writeError(w, http.StatusInternalServerError, "create failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, c)
}

func (h *CharacterHandler) updateCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	existing, err := h.repo.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "character not found")
		return
	}

	var req createCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Avatar != "" {
		existing.Avatar = req.Avatar
	}
	if req.Portrait != "" {
		existing.Portrait = req.Portrait
	}
	existing.Catchphrase = req.Catchphrase
	existing.Personality = req.Personality
	existing.Scenario = req.Scenario
	existing.FirstMessage = req.FirstMessage
	existing.ExampleDialogue = req.ExampleDialogue
	if req.Tags != nil {
		existing.Tags = req.Tags
	}
	if req.Scope != "" {
		existing.Scope = req.Scope
	}
	existing.RoomID = req.RoomID

	if err := h.repo.Update(id, existing); err != nil {
		writeError(w, http.StatusInternalServerError, "update failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

func (h *CharacterHandler) deleteCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.repo.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, "delete failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return time.Now().Format("20060102150405") + hex.EncodeToString(b)[:6]
}
