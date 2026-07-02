package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tavern-ai/backend/internal/database"
	"github.com/tavern-ai/backend/internal/models"
)

type SessionRepo struct {
	db *database.DB
}

func NewSessionRepo(db *database.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(s *models.Session) error {
	_, err := r.db.Exec(`
		INSERT INTO sessions (id, character_id, room_id, title, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		s.ID, s.CharacterID, s.RoomID, s.Title,
		s.CreatedAt.Format(time.RFC3339), s.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *SessionRepo) GetByID(id string) (*models.Session, error) {
	row := r.db.QueryRow(`SELECT id, character_id, room_id, title, created_at, updated_at
		FROM sessions WHERE id=?`, id)
	return scanSession(row)
}

func (r *SessionRepo) List() ([]models.Session, error) {
	rows, err := r.db.Query(`SELECT id, character_id, room_id, title, created_at, updated_at
		FROM sessions ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		s, err := scanSessionFromRows(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, *s)
	}
	return sessions, rows.Err()
}

func (r *SessionRepo) Update(id string, title string) error {
	_, err := r.db.Exec(`UPDATE sessions SET title=?, updated_at=? WHERE id=?`,
		title, time.Now().Format(time.RFC3339), id)
	return err
}

func (r *SessionRepo) Touch(id string) error {
	_, err := r.db.Exec(`UPDATE sessions SET updated_at=? WHERE id=?`,
		time.Now().Format(time.RFC3339), id)
	return err
}

func (r *SessionRepo) UpdateContext(id string, charID string, roomID *string) error {
	_, err := r.db.Exec(`UPDATE sessions SET character_id=?, room_id=?, updated_at=? WHERE id=?`,
		charID, roomID, time.Now().Format(time.RFC3339), id)
	return err
}

func (r *SessionRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE id=?", id)
	return err
}

func scanSession(scanner interface{ Scan(...any) error }) (*models.Session, error) {
	var (
		id, charID, title, createdAt, updatedAt string
		roomID                                  sql.NullString
	)
	err := scanner.Scan(&id, &charID, &roomID, &title, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	ct, _ := time.Parse(time.RFC3339, createdAt)
	ut, _ := time.Parse(time.RFC3339, updatedAt)
	s := &models.Session{
		ID: id, CharacterID: charID, Title: title,
		CreatedAt: ct, UpdatedAt: ut,
	}
	if roomID.Valid {
		s.RoomID = &roomID.String
	}
	return s, nil
}

func scanSessionFromRows(rows *sql.Rows) (*models.Session, error) {
	return scanSession(rows)
}

type MessageRepo struct {
	db *database.DB
}

func NewMessageRepo(db *database.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Create(m *models.Message) error {
	_, err := r.db.Exec(`
		INSERT INTO messages (id, session_id, role, content, timestamp)
		VALUES (?, ?, ?, ?, ?)`,
		m.ID, m.SessionID, m.Role, m.Content, m.Timestamp.Format(time.RFC3339),
	)
	return err
}

func (r *MessageRepo) ListBySession(sessionID string) ([]models.Message, error) {
	rows, err := r.db.Query(`SELECT id, session_id, role, content, timestamp
		FROM messages WHERE session_id=? ORDER BY timestamp ASC`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var msgs []models.Message
	for rows.Next() {
		var (
			id, sessID, role, content, ts string
		)
		if err := rows.Scan(&id, &sessID, &role, &content, &ts); err != nil {
			return nil, err
		}
		t, _ := time.Parse(time.RFC3339, ts)
		msgs = append(msgs, models.Message{
			ID: id, SessionID: sessID, Role: role, Content: content, Timestamp: t,
		})
	}
	return msgs, rows.Err()
}
