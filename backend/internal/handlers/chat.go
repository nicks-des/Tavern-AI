package handlers

import (
	"encoding/json"
	"fmt"
	"log"
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
	roomMsgRepo   *repository.RoomMessageRepo
	llmClient     *llm.Client
}

func NewChatHandler(
	sr *repository.SessionRepo,
	mr *repository.MessageRepo,
	cr *repository.CharacterRepo,
	wr *repository.WorldBookRepo,
	rr *repository.RoomRepo,
	rmr *repository.RoomMessageRepo,
	lc *llm.Client,
) *ChatHandler {
	return &ChatHandler{
		sessionRepo:   sr,
		messageRepo:   mr,
		characterRepo: cr,
		worldbookRepo: wr,
		roomRepo:      rr,
		roomMsgRepo:   rmr,
		llmClient:     lc,
	}
}

func (h *ChatHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/sessions/{id}/chat", h.handleChat)
	mux.HandleFunc("POST /api/rooms/{id}/run", h.handleRoomRun)
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

func compressHistory(client *llm.Client, msgs []models.Message) string {
	if len(msgs) < 6 || client == nil {
		return ""
	}
	oldMsgs := msgs[:len(msgs)-4] // all but last 4
	var lines []string
	for _, m := range oldMsgs {
		lines = append(lines, m.Content)
	}
	prompt := "Summarize this conversation in 1-2 short sentences: " + strings.Join(lines, "\n")
	resp, err := client.Chat([]llm.ChatMessage{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return ""
	}
	return resp
}

func updateWorldState(stateJSON, action, content, charName string) string {
	state := initWorldState()
	json.Unmarshal([]byte(stateJSON), state)
	state["lastAction"] = action
	state["lastActionBy"] = charName

	round := state["round"].(float64)
	state["round"] = round + 1

	// World clock: advance 5 minutes per round
	hr, min := parseClock(state["time"].(string))
	min += 5
	if min >= 60 { min -= 60; hr++ }
	if hr >= 24 { hr = 0; d := state["day"].(float64); state["day"] = d + 1 }
	state["time"] = fmt.Sprintf("%02d:%02d", hr, min)

	// Process ACT/REVEAL effects
	if action == "ACT" || action == "REVEAL" {
		events, _ := state["events"].([]any)
		t := state["time"].(string)
		state["events"] = append(events, fmt.Sprintf("[Day%d %s] %s: %s", int(state["day"].(float64)), t, charName, truncate(content, 60)))
	}

	contentLower := strings.ToLower(content)

	// Character status changes
	if strings.Contains(contentLower, "死") || strings.Contains(contentLower, "kill") || strings.Contains(contentLower, "杀") {
		setStatus(state, charName, "dying")
	}

	// LLM-driven relationships: parse [REL: char::other trust+1 fear-2]
	rels := getRelationships(state)
	for _, line := range []string{content, action + "|" + content} {
		for {
			start := strings.Index(line, "[REL:")
			if start < 0 { break }
			end := strings.Index(line[start:], "]") + start
			if end <= start { break }
			relStr := line[start+5 : end]
			processRelation(rels, relStr)
			line = line[end+1:]
		}
	}
	state["relationships"] = rels

	result, _ := json.Marshal(state)
	return string(result)
}

func parseClock(t string) (int, int) {
	h, m := 8, 0
	fmt.Sscanf(t, "%d:%d", &h, &m)
	return h, m
}

func processRelation(rels map[string]any, spec string) {
	// Format: char::other trust+1 fear-2
	parts := strings.SplitN(spec, " ", 2)
	if len(parts) != 2 { return }
	key := strings.TrimSpace(parts[0])
	mods := strings.TrimSpace(parts[1])

	if _, exists := rels[key]; !exists {
		rels[key] = map[string]any{"trust": float64(5), "fear": float64(0), "tag": ""}
	}
	r, ok := rels[key].(map[string]any)
	if !ok { return }

	for _, mod := range strings.Fields(mods) {
		if strings.HasPrefix(mod, "trust") {
			if v, err := fmt.Sscanf(mod, "trust%f", new(float64)); err == nil && v == 1 {
				r["trust"] = clampStat(r["trust"].(float64)+float64(0), 0, 10)
			}
		}
		if strings.HasPrefix(mod, "fear") {
			if v, err := fmt.Sscanf(mod, "fear%f", new(float64)); err == nil && v == 1 {
				r["fear"] = clampStat(r["fear"].(float64)+float64(0), 0, 10)
			}
		}
	}

	// Also handle explicit values: trust+2, fear-1
	for _, mod := range strings.Fields(mods) {
		val := float64(0)
		parsed := false
		if n := 0; strings.HasPrefix(mod, "trust+") {
			fmt.Sscanf(mod, "trust+%d", &n); val = float64(n); parsed = true
			r["trust"] = clampStat(r["trust"].(float64)+val, 0, 10)
		} else if strings.HasPrefix(mod, "trust-") {
			fmt.Sscanf(mod, "trust-%d", &n); val = -float64(n); parsed = true
			r["trust"] = clampStat(r["trust"].(float64)+val, 0, 10)
		} else if strings.HasPrefix(mod, "fear+") {
			fmt.Sscanf(mod, "fear+%d", &n); val = float64(n); parsed = true
			r["fear"] = clampStat(r["fear"].(float64)+val, 0, 10)
		} else if strings.HasPrefix(mod, "fear-") {
			fmt.Sscanf(mod, "fear-%d", &n); val = -float64(n); parsed = true
			r["fear"] = clampStat(r["fear"].(float64)+val, 0, 10)
		}
		_ = parsed
	}
}

func initWorldState() map[string]any {
	return map[string]any{
		"round":         float64(0),
		"time":          "08:00",
		"day":           float64(1),
		"relationships": make(map[string]any),
		"charStatus":    make(map[string]any),
		"events":        []any{},
		"lastAction":    "",
		"lastActionBy":  "",
	}
}

func getRelationships(state map[string]any) map[string]any {
	rels, ok := state["relationships"].(map[string]any)
	if !ok {
		return make(map[string]any)
	}
	if rels == nil {
		return make(map[string]any)
	}
	return rels
}

func setStatus(state map[string]any, key string, val any) {
	s, ok := state["charStatus"].(map[string]any)
	if !ok {
		s = make(map[string]any)
	}
	s[key] = val
	state["charStatus"] = s
}

func getCharStatus(state map[string]any, charID string) (string, bool) {
	s, ok := state["charStatus"].(map[string]any)
	if !ok {
		return "", false
	}
	// Search by name or ID
	for k, v := range s {
		if k == charID {
			return fmt.Sprint(v), true
		}
	}
	return "", false
}

func affectClosest(rels map[string]any, actor string, stat string, delta float64) {
	// Find the last person this actor interacted with and affect that relationship
	for k, v := range rels {
		if strings.Contains(k, actor+"::") {
			if r, ok := v.(map[string]any); ok {
				if cur, ok := r[stat].(float64); ok {
					r[stat] = clampStat(cur + delta, 0, 10)
				} else {
					r[stat] = clampStat(5+delta, 0, 10)
				}
				return
			}
		}
	}
}

func clampStat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}

func getWorldKey(state map[string]any, key string) any {
	if v, ok := state[key]; ok {
		return v
	}
	return "?"
}

func (h *ChatHandler) handleRoomRun(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("[RoomRun] panic: %v", rec)
			writeError(w, http.StatusInternalServerError, "internal error")
		}
	}()

	roomID := r.PathValue("id")
	log.Printf("[RoomRun] starting for room %s", roomID)
	if h.llmClient == nil {
		writeError(w, http.StatusServiceUnavailable, "LLM not available")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, _ := w.(http.Flusher)

	room, err := h.roomRepo.GetByID(roomID)
	if err != nil {
		fmt.Fprintf(w, "data: [ERROR] room not found\n\n")
		flusher.Flush()
		return
	}

	members, err := h.roomRepo.ListMembers(roomID)
	if err != nil || len(members) < 2 {
		fmt.Fprintf(w, "data: [ERROR] need at least 2 members\n\n")
		flusher.Flush()
		return
	}

	const maxRounds = 4 // auto-conversation stays at 4
	memberIdx := 0
	worldState := room.WorldState
	recentMessages := []models.Message{}

	for round := 0; round < maxRounds; round++ {
		select {
		case <-r.Context().Done():
			return
		default:
		}

		// Compress old messages if too many
		if len(recentMessages) > 10 {
			summary := compressHistory(h.llmClient, recentMessages)
			if summary != "" {
				recentMessages = []models.Message{
					{Role: "system", Content: "[Summary] " + summary},
				}
				recentMessages = append(recentMessages, recentMessages[len(recentMessages)-6:]...)
			}
		}

		// Round-robin through all members, skip dead/dying
		nextMember := &members[memberIdx%len(members)]
		memberIdx++

		// Check if character is dead
		ws := make(map[string]any)
		json.Unmarshal([]byte(worldState), &ws)
		if status, ok := getCharStatus(ws, nextMember.CharacterID); ok && (status == "dead") {
			continue
		}

		char, err := h.characterRepo.GetByID(nextMember.CharacterID)
		if err != nil {
			continue
		}

		// Decision
		decisionPrompt := fmt.Sprintf(
			`You are %s, goal: "%s", secret: "%s". World state: %s.
Conversation so far: %s
What do you do? SPEAK|content, ACT|content, or REVEAL|content. Keep it 1-2 sentences.`,
			char.Name, char.Goal, char.Secret, worldState, recentDialogue(recentMessages))

		dm := []llm.ChatMessage{
			{Role: "system", Content: "You are a character in a story. Respond with ACTION|content. Keep it brief."},
			{Role: "user", Content: decisionPrompt},
		}
		raw, err := h.llmClient.Chat(dm)
		if err != nil {
			continue
		}

		action := "SPEAK"
		content := raw
		if parts := strings.SplitN(raw, "|", 2); len(parts) == 2 {
			if a := strings.TrimSpace(parts[0]); a == "ACT" || a == "REVEAL" || a == "SPEAK" {
				action = a
				content = strings.TrimSpace(parts[1])
			}
		}

		if action == "ACT" || action == "REVEAL" {
			worldState = updateWorldState(worldState, action, content, char.Name)
			h.roomRepo.UpdateWorldState(roomID, worldState)
		}

		fmt.Fprintf(w, "data: [%s]\n\n", char.Name)
		flusher.Flush()

		execPrompt := fmt.Sprintf("As %s: %s", char.Name, content)
		msgs := h.llmClient.BuildMessages(char, recentMessages, execPrompt, "", room.WorldRules, "")
		resp, err := h.llmClient.ChatStream(msgs, func(token string) {
			fmt.Fprintf(w, "data: %s\n\n", token)
			flusher.Flush()
		})
		if err != nil {
			continue
		}

		recentMessages = append(recentMessages, models.Message{
			Role: "assistant", Content: resp,
		})
		if h.roomMsgRepo != nil {
			_ = h.roomMsgRepo.Create(&models.RoomMessage{
				ID:            generateID(),
				RoomID:        roomID,
				CharacterName: char.Name,
				Content:       resp,
				CreatedAt:     time.Now(),
			})
		}
	}

	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
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
		var ws2 map[string]any
		json.Unmarshal([]byte(worldState), &ws2)

		// Step 1: Decision prompt
		type autoDecision struct {
			Action  string `json:"action"`  // SPEAK / ACT / REVEAL
			Content string `json:"content"` // what to say or do
		}

		decisionPrompt := fmt.Sprintf(
			`You are %s, with goal: "%s", secret: "%s".
World: Time %v | Round %.0f | Status: %v
Current relationships: %v

Recent conversation:
%s

Choose action: SPEAK | ACT | REVEAL
Format: ACTION|your response

When your action affects relationships, add [REL: Actr::Target trust+2] or [REL: Actr::Target fear-1]
Actions with "kill/die/杀死/死亡" will mark characters as dying.`, char.Name, char.Goal, char.Secret,
				getWorldKey(ws2, "time"), getWorldKey(ws2, "round"), ws2["charStatus"], ws2["relationships"], recentDialogue(recentMessages))

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
			newState := updateWorldState(worldState, decision.Action, decision.Content, char.Name)
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
