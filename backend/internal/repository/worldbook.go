package repository

import (
	"time"

	"github.com/tavern-ai/backend/internal/database"
	"github.com/tavern-ai/backend/internal/models"
)

type WorldBookRepo struct {
	db *database.DB
}

func NewWorldBookRepo(db *database.DB) *WorldBookRepo {
	return &WorldBookRepo{db: db}
}

func (r *WorldBookRepo) Create(entry *models.WorldBookEntry) error {
	enabled := 0
	if entry.Enabled {
		enabled = 1
	}
	_, err := r.db.Exec(`
		INSERT INTO world_book_entries (id, character_id, keywords, content, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.CharacterID, entry.Keywords, entry.Content,
		enabled, entry.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *WorldBookRepo) GetByID(id string) (*models.WorldBookEntry, error) {
	var e models.WorldBookEntry
	var en int
	var ts string
	err := r.db.QueryRow(`
		SELECT id, character_id, keywords, content, enabled, created_at
		FROM world_book_entries WHERE id=?`, id).Scan(&e.ID, &e.CharacterID, &e.Keywords, &e.Content, &en, &ts)
	if err != nil {
		return nil, err
	}
	e.Enabled = en == 1
	e.CreatedAt, _ = time.Parse(time.RFC3339, ts)
	return &e, nil
}

func (r *WorldBookRepo) ListByCharacter(charID string) ([]models.WorldBookEntry, error) {
	rows, err := r.db.Query(`
		SELECT id, character_id, keywords, content, enabled, created_at
		FROM world_book_entries WHERE character_id=? ORDER BY created_at ASC`, charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.WorldBookEntry
	for rows.Next() {
		var e models.WorldBookEntry
		var en int
		var ts string
		if err := rows.Scan(&e.ID, &e.CharacterID, &e.Keywords, &e.Content, &en, &ts); err != nil {
			return nil, err
		}
		e.Enabled = en == 1
		e.CreatedAt, _ = time.Parse(time.RFC3339, ts)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (r *WorldBookRepo) Update(entry *models.WorldBookEntry) error {
	enabled := 0
	if entry.Enabled {
		enabled = 1
	}
	_, err := r.db.Exec(`
		UPDATE world_book_entries SET keywords=?, content=?, enabled=?
		WHERE id=?`,
		entry.Keywords, entry.Content, enabled, entry.ID,
	)
	return err
}

func (r *WorldBookRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM world_book_entries WHERE id=?", id)
	return err
}

func (r *WorldBookRepo) GetEnabledByCharacter(charID string) ([]models.WorldBookEntry, error) {
	rows, err := r.db.Query(`
		SELECT id, character_id, keywords, content, enabled, created_at
		FROM world_book_entries WHERE character_id=? AND enabled=1 ORDER BY created_at ASC`, charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.WorldBookEntry
	for rows.Next() {
		var e models.WorldBookEntry
		var en int
		var ts string
		if err := rows.Scan(&e.ID, &e.CharacterID, &e.Keywords, &e.Content, &en, &ts); err != nil {
			return nil, err
		}
		e.Enabled = en == 1
		e.CreatedAt, _ = time.Parse(time.RFC3339, ts)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
