package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/tavern-ai/backend/internal/models"
	"github.com/tavern-ai/backend/internal/repository"
)

type RoomHandler struct {
	repo      *repository.RoomRepo
	charRepo  *repository.CharacterRepo
}

func NewRoomHandler(repo *repository.RoomRepo, charRepo *repository.CharacterRepo) *RoomHandler {
	return &RoomHandler{repo: repo, charRepo: charRepo}
}

func (h *RoomHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/rooms", h.listRooms)
	mux.HandleFunc("POST /api/rooms", h.createRoom)
	mux.HandleFunc("GET /api/rooms/{id}", h.getRoom)
	mux.HandleFunc("PUT /api/rooms/{id}", h.updateRoom)
	mux.HandleFunc("DELETE /api/rooms/{id}", h.deleteRoom)

	mux.HandleFunc("POST /api/rooms/{id}/members", h.addMember)
	mux.HandleFunc("DELETE /api/rooms/{id}/members/{charId}", h.removeMember)
}

func (h *RoomHandler) listRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.repo.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rooms == nil {
		rooms = []models.Room{}
	}
	writeJSON(w, http.StatusOK, rooms)
}

func (h *RoomHandler) createRoom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		WorldRules  string `json:"worldRules"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	now := time.Now()
	room := &models.Room{
		ID:          generateID(),
		Name:        req.Name,
		Description: req.Description,
		WorldRules:  req.WorldRules,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := h.repo.Create(room); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, room)
}

func (h *RoomHandler) getRoom(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	room, err := h.repo.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "room not found")
		return
	}

	members, _ := h.repo.ListMembers(id)
	if members == nil {
		members = []models.RoomMember{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"room":    room,
		"members": members,
	})
}

func (h *RoomHandler) updateRoom(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	room, err := h.repo.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "room not found")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		WorldRules  string `json:"worldRules"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Name != "" {
		room.Name = req.Name
	}
	room.Description = req.Description
	room.WorldRules = req.WorldRules

	if err := h.repo.Update(room); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, room)
}

func (h *RoomHandler) deleteRoom(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.repo.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *RoomHandler) addMember(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")

	var req struct {
		CharacterID string `json:"characterId"`
		Overrides   string `json:"overrides"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.CharacterID == "" {
		writeError(w, http.StatusBadRequest, "characterId is required")
		return
	}

	if req.Overrides == "" {
		req.Overrides = "{}"
	}

	if err := h.repo.AddMember(roomID, req.CharacterID, req.Overrides); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "added"})
}

func (h *RoomHandler) removeMember(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")
	charID := r.PathValue("charId")

	if err := h.repo.RemoveMember(roomID, charID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}
