package repository

import (
	"fmt"
	"time"

	"github.com/tavern-ai/backend/internal/database"
	"github.com/tavern-ai/backend/internal/models"
)

type RoomMessageRepo struct {
	db *database.DB
}

func NewRoomMessageRepo(db *database.DB) *RoomMessageRepo {
	r := &RoomMessageRepo{db: db}
	r.init()
	return r
}

func (r *RoomMessageRepo) init() {
	r.db.Exec(`CREATE TABLE IF NOT EXISTS room_messages (
		id TEXT PRIMARY KEY,
		room_id TEXT NOT NULL,
		character_name TEXT DEFAULT '',
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
	)`)
}

func (r *RoomMessageRepo) Create(msg *models.RoomMessage) error {
	_, err := r.db.Exec(`
		INSERT INTO room_messages (id, room_id, character_name, content, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		msg.ID, msg.RoomID, msg.CharacterName, msg.Content, msg.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *RoomMessageRepo) ListByRoom(roomID string) ([]models.RoomMessage, error) {
	rows, err := r.db.Query(`SELECT id, room_id, character_name, content, created_at
		FROM room_messages WHERE room_id=? ORDER BY created_at ASC`, roomID)
	if err != nil {
		return nil, fmt.Errorf("list room messages: %w", err)
	}
	defer rows.Close()

	var msgs []models.RoomMessage
	for rows.Next() {
		var m models.RoomMessage
		var ts string
		if err := rows.Scan(&m.ID, &m.RoomID, &m.CharacterName, &m.Content, &ts); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, ts)
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (r *RoomMessageRepo) DeleteByRoom(roomID string) error {
	_, err := r.db.Exec("DELETE FROM room_messages WHERE room_id=?", roomID)
	return err
}
