package models

import "time"

type CharacterScope string

const (
	ScopeGlobal CharacterScope = "global"
	ScopeRoom   CharacterScope = "room"
)

type Character struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Avatar          string         `json:"avatar"`
	Portrait        string         `json:"portrait"`
	Catchphrase     string         `json:"catchphrase"`
	Description     string         `json:"description"`
	Personality     string         `json:"personality"`
	Scenario        string         `json:"scenario"`
	FirstMessage    string         `json:"firstMessage"`
	ExampleDialogue string         `json:"exampleDialogue"`
	Goal           string         `json:"goal"`
	Secret         string         `json:"secret"`
	Tags            []string       `json:"tags"`
	Scope           CharacterScope `json:"scope"`
	RoomID          *string        `json:"roomId"`
	CreatedAt       time.Time      `json:"createdAt"`
}

type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	Role      string    `json:"role"` // "user" | "assistant" | "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Session struct {
	ID          string    `json:"id"`
	CharacterID string    `json:"characterId"`
	RoomID      *string   `json:"roomId"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Room struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	WorldRules  string       `json:"worldRules"`
	WorldState  string       `json:"worldState"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

type RoomMember struct {
	RoomID      string `json:"roomId"`
	CharacterID string `json:"characterId"`
	// Overrides are stored as JSON string in SQLite
	Overrides string `json:"overrides"`
}

type WorldBookEntry struct {
	ID          string    `json:"id"`
	CharacterID string    `json:"characterId"`
	Keywords    string    `json:"keywords"`
	Content     string    `json:"content"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
}
