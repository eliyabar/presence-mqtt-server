package services

import (
	"database/sql"
	"fmt"
)

type PresenceService struct {
	db                        *sql.DB
	userPresenceByUserIdCache map[int]bool
}

type Presence struct {
	UserID     int
	IsPresent  bool
	Name       string
	AvatarName string
	LastUpdate string
}

func createPresenceService(db *sql.DB) *PresenceService {
	ps := &PresenceService{db: db,
		userPresenceByUserIdCache: make(map[int]bool)}

	presence, err := ps.GetPresence()
	if err != nil {
		fmt.Println("could not fill cache from db")
	}
	for _, p := range presence {
		ps.userPresenceByUserIdCache[p.UserID] = p.IsPresent
	}
	return ps
}

func (ps *PresenceService) IsPresent(userId int) bool {
	if ps.userPresenceByUserIdCache == nil {
		return false
	} else {
		if _, ok := ps.userPresenceByUserIdCache[userId]; ok {
			return true
		}
		return false
	}
}

func (ps *PresenceService) UpsertPresence(userId int, isPresent bool) (bool, error) {
	if v, ok := ps.userPresenceByUserIdCache[userId]; ok && v == isPresent {
		fmt.Println("value is cached")
		return false, nil
	}
	insertPresenceSQL := `INSERT INTO presence (user_id, is_present) VALUES ($1, $2) ON CONFLICT(user_id) DO UPDATE SET is_present = $2, last_update = CURRENT_TIMESTAMP`
	statement, err := ps.db.Prepare(insertPresenceSQL)

	if err != nil {
		return false, fmt.Errorf("could not prepare query %w", err)

	}
	_, err = statement.Exec(userId, isPresent)
	if err != nil {
		return false, fmt.Errorf("could not execute query %w", err)

	}
	ps.userPresenceByUserIdCache[userId] = isPresent

	return true, nil
}

func (ps *PresenceService) ResetPresence() error {
	insertPresenceSQL := `DELETE FROM presence`
	statement, err := ps.db.Prepare(insertPresenceSQL)
	ps.userPresenceByUserIdCache = make(map[int]bool)
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
