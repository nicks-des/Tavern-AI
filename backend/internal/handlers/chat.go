package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tavern-ai/backend/internal/llm"
	"github.com/tavern-ai/backend/internal/models"
	"github.com/tavern-ai/backend/internal/repository"
)

type ChatHandler struct {
	sessionRepo   *repository.SessionRepo
	messageRepo   *repository.MessageRepo
	characterRepo *repository.CharacterRepo
	worldbookRepo *repository.WorldBookRepo
	roomRepo      *repository.RoomRepo
	llmClient     *llm.Client
}

func NewChatHandler(
	sr *repository.SessionRepo,
	mr *repository.MessageRepo,
	cr *repository.CharacterRepo,
	wr *repository.WorldBookRepo,
	rr *repository.RoomRepo,
	lc *llm.Client,
) *ChatHandler {
	return &ChatHandler{
		sessionRepo:   sr,
		messageRepo:   mr,
		characterRepo: cr,
		worldbookRepo: wr,
		roomRepo:      rr,
		llmClient:     lc,
	}
}

func (h *ChatHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/sessions/{id}/chat", h.handleChat)
}

func (h *ChatHandler) handleChat(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	var req struct {
		Message     string  `json:"message"`
		CharacterID string  `json:"characterId"`
		RoomID      *string `json:"roomId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	// Update session with room/character info if provided
	if req.RoomID != nil || req.CharacterID != "" {
		h.sessionRepo.UpdateContext(sessionID, req.CharacterID, req.RoomID)
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

	worldbookContext := ""
	if h.worldbookRepo != nil {
		entries, err := h.worldbookRepo.GetEnabledByCharacter(session.CharacterID)
		if err == nil {
			for _, entry := range entries {
				if matchKeywords(req.Message, entry.Keywords) {
					worldbookContext += entry.Content + "\n"
				}
			}
		}
	}

	roomContext := ""
	roomOverrides := ""
	if session.RoomID != nil && h.roomRepo != nil {
		room, err := h.roomRepo.GetByID(*session.RoomID)
		if err == nil && room.WorldRules != "" {
			roomContext = room.WorldRules
		}

		members, _ := h.roomRepo.ListMembers(*session.RoomID)
		for _, m := range members {
			if m.CharacterID == session.CharacterID && m.Overrides != "" {
				roomOverrides = m.Overrides
			}
		}
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

	messages := h.llmClient.BuildMessages(character, history, req.Message, worldbookContext, roomContext, roomOverrides)

	if h.llmClient == nil {
		h.handleMockChat(w, sessionID, character, req.Message)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
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

	// Auto-conversation for room members
	if session.RoomID != nil && h.roomRepo != nil && h.llmClient != nil {
		h.runAutoConversation(w, flusher, session, character, history, aiMsg, roomContext)
	}
}

func (h *ChatHandler) handleMockChat(w http.ResponseWriter, sessionID string, character *models.Character, msg string) {
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
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

func matchKeywords(message, keywords string) bool {
	message = strings.ToLower(message)
	for _, kw := range strings.Split(keywords, ",") {
		kw = strings.TrimSpace(strings.ToLower(kw))
		if kw != "" && strings.Contains(message, kw) {
			return true
		}
	}
	return false
}

func recentDialogue(msgs []models.Message) string {
	var lines []string
	start := len(msgs) - 6
	if start < 0 {
		start = 0
	}
	for _, m := range msgs[start:] {
		lines = append(lines, m.Content)
	}
	return strings.Join(lines, "\n")
}

func updateWorldState(stateJSON, action, content string) string {
	state := make(map[string]any)
	json.Unmarshal([]byte(stateJSON), &state)
	state["lastAction"] = action
	state["lastActionContent"] = content
	state["timestamp"] = time.Now().Unix()
	result, _ := json.Marshal(state)
	return string(result)
}

func (h *ChatHandler) runAutoConversation(w http.ResponseWriter, flusher http.Flusher, session *models.Session, currentChar *models.Character, history []models.Message, lastMsg *models.Message, roomContext string) {
	const maxRounds = 4
	rMsgs := make([]models.Message, len(history))
	copy(rMsgs, history)
	recentMessages := append(rMsgs, *lastMsg)

	members, err := h.roomRepo.ListMembers(*session.RoomID)
	if err != nil || len(members) <= 1 {
		return
	}

	lastSpeakerID := session.CharacterID
	for round := 0; round < maxRounds; round++ {
		var nextMember *models.RoomMember
		for _, m := range members {
			if m.CharacterID != lastSpeakerID {
				nextMember = &m
				break
			}
		}
		for _, m := range members {
			if m.CharacterID != lastSpeakerID && m.CharacterID != nextMember.CharacterID {
				nextMember = &m
				break
			}
		}

		if nextMember == nil || nextMember.CharacterID == lastSpeakerID {
			break
		}

		char, err := h.characterRepo.GetByID(nextMember.CharacterID)
		if err != nil {
			continue
		}

		wbContext := ""
		if h.worldbookRepo != nil {
			entries, _ := h.worldbookRepo.GetEnabledByCharacter(char.ID)
			for _, e := range entries {
				lastContent := lastMsg.Content
				for _, rm := range recentMessages {
					lastContent = rm.Content
				}
				_ = lastContent
				if len(e.Keywords) > 0 {
					for _, rm := range recentMessages {
						if matchKeywords(rm.Content, e.Keywords) {
							wbContext += e.Content + "\n"
							break
						}
					}
				}
			}
		}

		charOverrides := ""
		for _, m := range members {
			if m.CharacterID == char.ID && m.Overrides != "" {
				charOverrides = m.Overrides
			}
		}

		// Load world state
		worldState := "{}"
		if room, err := h.roomRepo.GetByID(*session.RoomID); err == nil {
			worldState = room.WorldState
		}

		// Step 1: Decision prompt
		type autoDecision struct {
			Action  string `json:"action"`  // SPEAK / ACT / REVEAL
			Content string `json:"content"` // what to say or do
		}

		decisionPrompt := fmt.Sprintf(
			`You are %s, with goal: "%s", secret: "%s".
The current world state is: %s

Recent conversation:
%s

Given your personality, goal, and the current situation, what do you want to do?
Respond with ONLY:
- SPEAK: just say something in character
- ACT: take an action that changes the world (e.g. leave the room, draw a weapon) - describe the action
- REVEAL: share a secret you've been hiding

Then write your response. For ACT/REVEAL, the world state will update automatically.
Format: ACTION|content`, char.Name, char.Goal, char.Secret, worldState, recentDialogue(recentMessages))

		decisionMessages := []llm.ChatMessage{
			{Role: "system", Content: "You are a character making a decision. Keep it brief, 1-2 sentences. Format your response as ACTION|content where ACTION is SPEAK, ACT, or REVEAL."},
			{Role: "user", Content: decisionPrompt},
		}

		var decision autoDecision
		decision.Action = "SPEAK"
		rawDecision, err := h.llmClient.Chat(decisionMessages)
		if err == nil {
			parts := strings.SplitN(rawDecision, "|", 2)
			if len(parts) == 2 {
				action := strings.TrimSpace(parts[0])
				content := strings.TrimSpace(parts[1])
				if action == "ACT" || action == "REVEAL" || action == "SPEAK" {
					decision.Action = action
					decision.Content = content
				}
			} else {
				decision.Content = rawDecision
			}
		}

		// Step 2: Execute decision - update world state if needed
		if decision.Action == "ACT" || decision.Action == "REVEAL" {
			newState := updateWorldState(worldState, decision.Action, decision.Content)
			h.roomRepo.UpdateWorldState(*session.RoomID, newState)
		}

		// Step 3: Send character name tag + generate response
		fmt.Fprintf(w, "data: [%s]\n\n", char.Name)
		flusher.Flush()

		execPrompt := fmt.Sprintf(
			"As %s, respond to the recent conversation. Your goal: %s. Secret: %s. World: %s. %s",
			char.Name, char.Goal, char.Secret, worldState, decision.Content,
		)

		messages := h.llmClient.BuildMessages(char, recentMessages, execPrompt, wbContext, roomContext, charOverrides)

		fmt.Fprintf(w, "data: [%s]\n\n", char.Name)
		flusher.Flush()

		response, err := h.llmClient.ChatStream(messages, func(token string) {
			fmt.Fprintf(w, "data: %s\n\n", token)
			flusher.Flush()
		})

		if err != nil {
			continue
		}

		autoMsg := &models.Message{
			ID:        generateID(),
			SessionID: session.ID,
			Role:      "assistant",
			Content:   response,
			Timestamp: time.Now(),
		}
		h.messageRepo.Create(autoMsg)
		recentMessages = append(recentMessages, *autoMsg)
		lastSpeakerID = char.ID
		lastMsg = autoMsg
	}
}
