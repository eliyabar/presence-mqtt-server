package services

import (
	"database/sql"
	"fmt"
	"log"
)

type UserService struct {
	db           *sql.DB
	userMacCache map[string]User
}

type User struct {
	UID        int
	Name       string
	AvatarName string
	MacAddress string
	Created    string
}

func createUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}
func (us *UserService) CreateUser(name, avatarUrl, macAddress string) error {
	insertUserSQL := `INSERT INTO users (name, avatar_name, mac_address) VALUES ($1, $2, $3) ON CONFLICT DO UPDATE SET name =$1, avatar_name=$2`
	statement, err := us.db.Prepare(insertUserSQL)
	if err != nil {
		return fmt.Errorf("could not prepare query %w", err)
	}
	_, err = statement.Exec(name, avatarUrl, macAddress)
	if err != nil {
		return fmt.Errorf("could not execute query %w", err)
	}
	// reset cache
	us.userMacCache = nil

	return nil
}

func (us *UserService) GetUserByMac(mac string) (*User, error) {
	if us.userMacCache == nil {
		us.userMacCache = map[string]User{}

		users, err := us.GetUsers()
		if err != nil {
			return nil, fmt.Errorf("could not get users in GetUserByMac %w", err)
		}

		for _, user := range users {
			us.userMacCache[user.MacAddress] = user
		}
	}
	if user, ok := us.userMacCache[mac]; ok {
		return &user, nil
	}
	return nil, fmt.Errorf("could not find mac address")
}

func (us *UserService) GetUsers() ([]User, error) {
	var users []User
	row, err := us.db.Query("SELECT * FROM users")

	if err != nil {
		return nil, fmt.Errorf("could not fetch users %w", err)
	}
	defer row.Close()
	for row.Next() {
		var uid int
		var name string
		var avatar_name string
		var mac_address string
		var created string
		err := row.Scan(&uid, &name, &avatar_name, &mac_address, &created)
		if err != nil {
			log.Fatal(err)
		}
		user := User{
			UID:        uid,
			Name:       name,
			AvatarName: avatar_name,
			MacAddress: mac_address,
			Created:    created,
		}
		users = append(users, user)
	}
	return users, nil
}
