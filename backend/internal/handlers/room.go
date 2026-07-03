package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tavern-ai/backend/internal/models"
	"github.com/tavern-ai/backend/internal/repository"
)

type RoomHandler struct {
	repo      *repository.RoomRepo
	charRepo  *repository.CharacterRepo
	msgRepo   *repository.RoomMessageRepo
}

func NewRoomHandler(repo *repository.RoomRepo, charRepo *repository.CharacterRepo, msgRepo *repository.RoomMessageRepo) *RoomHandler {
	return &RoomHandler{repo: repo, charRepo: charRepo, msgRepo: msgRepo}
}

func (h *RoomHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/rooms", h.listRooms)
	mux.HandleFunc("POST /api/rooms", h.createRoom)
	mux.HandleFunc("GET /api/rooms/{id}", h.getRoom)
	mux.HandleFunc("PUT /api/rooms/{id}", h.updateRoom)
	mux.HandleFunc("DELETE /api/rooms/{id}", h.deleteRoom)

	mux.HandleFunc("POST /api/rooms/{id}/members", h.addMember)
	mux.HandleFunc("DELETE /api/rooms/{id}/members/{charId}", h.removeMember)

	// Room messages history
	if h.msgRepo != nil {
		mux.HandleFunc("GET /api/rooms/{id}/messages", h.listMessages)
		mux.HandleFunc("GET /api/rooms/{id}/export", h.exportMessages)
		mux.HandleFunc("DELETE /api/rooms/{id}/messages", h.resetWorld)
	}
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
		WorldState:  `{"round":0,"relationships":{},"charStatus":{},"events":[]}`,
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

func (h *RoomHandler) listMessages(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	msgs, err := h.msgRepo.ListByRoom(id)
	if err != nil {
		msgs = []models.RoomMessage{}
	}
	if msgs == nil {
		msgs = []models.RoomMessage{}
	}
	writeJSON(w, http.StatusOK, msgs)
}

func (h *RoomHandler) exportMessages(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	room, err := h.repo.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "room not found")
		return
	}

	msgs, _ := h.msgRepo.ListByRoom(id)

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("# %s\n\n", room.Name))
	buf.WriteString(fmt.Sprintf("> %s\n\n", room.Description))

	if room.WorldRules != "" {
		buf.WriteString(fmt.Sprintf("**世界规则**: %s\n\n", room.WorldRules))
	}
	if room.WorldState != "{}" && room.WorldState != `{"round":0}` {
		buf.WriteString(fmt.Sprintf("**世界状态**: `%s`\n\n", room.WorldState))
	}
	buf.WriteString("---\n\n")

	for _, m := range msgs {
		buf.WriteString(fmt.Sprintf("**%s**: %s\n\n", m.CharacterName, m.Content))
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.md"`, room.Name))
	w.Write([]byte(buf.String()))
}

func (h *RoomHandler) resetWorld(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.msgRepo.DeleteByRoom(id)
	h.repo.UpdateWorldState(id, `{"round":0,"relationships":{},"charStatus":{},"events":[]}`)
	writeJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}
