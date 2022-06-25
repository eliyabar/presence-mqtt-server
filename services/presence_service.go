package services

import (
	"database/sql"
	"fmt"
)

type PresenceService struct {
	db *sql.DB
}

type Presence struct {
	UserID     int
	IsPresent  bool
	Name       string
	AvatarName string
	LastUpdate string
}

func createPresenceService(db *sql.DB) *PresenceService {
	return &PresenceService{db: db}
}

func (ps *PresenceService) UpsertPresence(userId int, isPresent bool) error {
	insertPresenceSQL := `INSERT INTO presence (user_id, is_present) VALUES ($1, $2) ON CONFLICT(user_id) DO UPDATE SET is_present = $2, last_update = CURRENT_TIMESTAMP`
	statement, err := ps.db.Prepare(insertPresenceSQL)

	if err != nil {
		return fmt.Errorf("could not prepare query %w", err)

	}
	_, err = statement.Exec(userId, isPresent)
	if err != nil {
		return fmt.Errorf("could not execute query %w", err)

	}
	return nil
}

func (ps *PresenceService) ResetPresence() error {
	insertPresenceSQL := `DELETE FROM presence`
	statement, err := ps.db.Prepare(insertPresenceSQL)

	if err != nil {
		return fmt.Errorf("could not prepare query %w", err)

	}
	_, err = statement.Exec()
	if err != nil {
		return fmt.Errorf("could not execute query %w", err)

	}
	return nil
}

func (ps *PresenceService) GetPresence() ([]Presence, error) {
	var presence []Presence
	row, err := ps.db.Query("SELECT user_id, is_present, name, avatar_name, last_update from presence INNER JOIN users u on u.uid = presence.user_id")
	if err != nil {
		return nil, fmt.Errorf("could not query %w", err)
	}
	defer row.Close()

	for row.Next() {
		var user_id int
		var is_present bool
		var name string
		var avatar_name string
		var last_update string
		err := row.Scan(&user_id, &is_present, &name, &avatar_name, &last_update)

		if err != nil {
			return nil, fmt.Errorf("could not scan row %w", err)
		}
		p := Presence{
			UserID:     user_id,
			IsPresent:  is_present,
			Name:       name,
			AvatarName: avatar_name,
			LastUpdate: last_update,
		}
		presence = append(presence, p)
	}

	return presence, nil
}