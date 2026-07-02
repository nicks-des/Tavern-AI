package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tavern-ai/backend/internal/database"
	"github.com/tavern-ai/backend/internal/models"
)

type CharacterRepo struct {
	db *database.DB
}

func NewCharacterRepo(db *database.DB) *CharacterRepo {
	return &CharacterRepo{db: db}
}

func (r *CharacterRepo) Create(c *models.Character) error {
	tagsJSON, _ := json.Marshal(c.Tags)
	_, err := r.db.Exec(`
		INSERT INTO characters (id, name, avatar, portrait, catchphrase, description,
			personality, scenario, first_message, example_dialogue, goal, secret, tags, scope, room_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.Name, c.Avatar, c.Portrait, c.Catchphrase, c.Description,
		c.Personality, c.Scenario, c.FirstMessage, c.ExampleDialogue,
		c.Goal, c.Secret, string(tagsJSON), c.Scope, c.RoomID, c.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *CharacterRepo) GetByID(id string) (*models.Character, error) {
	row := r.db.QueryRow(`SELECT id, name, avatar, portrait, catchphrase, description,
		personality, scenario, first_message, example_dialogue, goal, secret, tags, scope, room_id, created_at
		FROM characters WHERE id=?`, id)
	return scanCharacter(row)
}

func (r *CharacterRepo) ListGlobal() ([]models.Character, error) {
	return r.query(`SELECT id, name, avatar, portrait, catchphrase, description,
		personality, scenario, first_message, example_dialogue, goal, secret, tags, scope, room_id, created_at
		FROM characters WHERE scope='global' ORDER BY created_at DESC`)
}

func (r *CharacterRepo) ListByRoom(roomID string) ([]models.Character, error) {
	return r.query(`SELECT id, name, avatar, portrait, catchphrase, description,
		personality, scenario, first_message, example_dialogue, goal, secret, tags, scope, room_id, created_at
		FROM characters WHERE scope='room' AND room_id=? ORDER BY created_at DESC`, roomID)
}

func (r *CharacterRepo) Update(id string, c *models.Character) error {
	tagsJSON, _ := json.Marshal(c.Tags)
	_, err := r.db.Exec(`
		UPDATE characters SET name=?, avatar=?, portrait=?, catchphrase=?, description=?,
		personality=?, scenario=?, first_message=?, example_dialogue=?, goal=?, secret=?, tags=?,
		scope=?, room_id=?
		WHERE id=?`,
		c.Name, c.Avatar, c.Portrait, c.Catchphrase, c.Description,
		c.Personality, c.Scenario, c.FirstMessage, c.ExampleDialogue,
		c.Goal, c.Secret, string(tagsJSON), c.Scope, c.RoomID, id,
	)
	return err
}

func (r *CharacterRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM characters WHERE id=?", id)
	return err
}

func (r *CharacterRepo) query(query string, args ...any) ([]models.Character, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query characters: %w", err)
	}
	defer rows.Close()

	var chars []models.Character
	for rows.Next() {
		c, err := scanCharacterFromRows(rows)
		if err != nil {
			return nil, err
		}
		chars = append(chars, *c)
	}
	return chars, rows.Err()
}

func scanCharacter(scanner interface{ Scan(...any) error }) (*models.Character, error) {
	var (
		id, name, avatar, portrait, catchphrase, desc, personality, scenario string
		firstMsg, example, goal, secret, tagsJSON, scope                    string
		roomID                                                              sql.NullString
		createdAt                                                           string
	)
	err := scanner.Scan(&id, &name, &avatar, &portrait, &catchphrase, &desc,
		&personality, &scenario, &firstMsg, &example, &goal, &secret,
		&tagsJSON, &scope, &roomID, &createdAt)
	if err != nil {
		return nil, err
	}

	var tags []string
	json.Unmarshal([]byte(tagsJSON), &tags)
	if tags == nil {
		tags = []string{}
	}

	t, _ := time.Parse(time.RFC3339, createdAt)

	c := &models.Character{
		ID: id, Name: name, Avatar: avatar, Portrait: portrait,
		Catchphrase: catchphrase, Description: desc, Personality: personality,
		Scenario: scenario, FirstMessage: firstMsg, ExampleDialogue: example,
		Goal: goal, Secret: secret,
		Tags: tags, Scope: models.CharacterScope(scope), CreatedAt: t,
	}
	if roomID.Valid {
		c.RoomID = &roomID.String
	}
	return c, nil
}

func scanCharacterFromRows(rows *sql.Rows) (*models.Character, error) {
	return scanCharacter(rows)
}
