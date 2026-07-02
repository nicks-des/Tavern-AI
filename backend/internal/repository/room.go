package repository

import (
	"time"

	"github.com/tavern-ai/backend/internal/database"
	"github.com/tavern-ai/backend/internal/models"
)

type RoomRepo struct {
	db *database.DB
}

func NewRoomRepo(db *database.DB) *RoomRepo {
	return &RoomRepo{db: db}
}

func (r *RoomRepo) Create(room *models.Room) error {
	_, err := r.db.Exec(`
		INSERT INTO rooms (id, name, description, world_rules, world_state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		room.ID, room.Name, room.Description, room.WorldRules, room.WorldState,
		room.CreatedAt.Format(time.RFC3339), room.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *RoomRepo) GetByID(id string) (*models.Room, error) {
	var room models.Room
	var ca, ua string
	err := r.db.QueryRow(`
		SELECT id, name, description, world_rules, world_state, created_at, updated_at
		FROM rooms WHERE id=?`, id).Scan(&room.ID, &room.Name, &room.Description, &room.WorldRules, &room.WorldState, &ca, &ua)
	if err != nil {
		return nil, err
	}
	room.CreatedAt, _ = time.Parse(time.RFC3339, ca)
	room.UpdatedAt, _ = time.Parse(time.RFC3339, ua)
	return &room, nil
}

func (r *RoomRepo) List() ([]models.Room, error) {
	rows, err := r.db.Query(`SELECT id, name, description, world_rules, world_state, created_at, updated_at
		FROM rooms ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var room models.Room
		var ca, ua string
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.WorldRules, &room.WorldState, &ca, &ua); err != nil {
			return nil, err
		}
		room.CreatedAt, _ = time.Parse(time.RFC3339, ca)
		room.UpdatedAt, _ = time.Parse(time.RFC3339, ua)
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}

func (r *RoomRepo) Update(room *models.Room) error {
	_, err := r.db.Exec(`
		UPDATE rooms SET name=?, description=?, world_rules=?, updated_at=?
		WHERE id=?`,
		room.Name, room.Description, room.WorldRules, time.Now().Format(time.RFC3339), room.ID,
	)
	return err
}

func (r *RoomRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM rooms WHERE id=?", id)
	return err
}

func (r *RoomRepo) AddMember(roomID, characterID, overrides string) error {
	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO room_members (room_id, character_id, overrides)
		VALUES (?, ?, ?)`, roomID, characterID, overrides)
	return err
}

func (r *RoomRepo) UpdateWorldState(id string, state string) error {
	_, err := r.db.Exec(`UPDATE rooms SET world_state=?, updated_at=? WHERE id=?`,
		state, time.Now().Format(time.RFC3339), id)
	return err
}

func (r *RoomRepo) RemoveMember(roomID, characterID string) error {
	_, err := r.db.Exec("DELETE FROM room_members WHERE room_id=? AND character_id=?", roomID, characterID)
	return err
}

func (r *RoomRepo) ListMembers(roomID string) ([]models.RoomMember, error) {
	rows, err := r.db.Query(`SELECT room_id, character_id, overrides FROM room_members WHERE room_id=?`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.RoomMember
	for rows.Next() {
		var m models.RoomMember
		if err := rows.Scan(&m.RoomID, &m.CharacterID, &m.Overrides); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}
