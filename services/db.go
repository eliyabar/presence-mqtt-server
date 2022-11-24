package services

import (
	"database/sql"
	"fmt"
	"github.com/spf13/viper"
	"sync"
)

var lock = &sync.Mutex{}

type DB struct {
	PresenceService *PresenceService
	UserService     *UserService
	db              *sql.DB
}

func (d *DB) CloseConnection() {
	err := d.db.Close()
	if err != nil {
		return
	}
}

var dbInstance *DB

func GetDBInstance() *DB {
	if dbInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if dbInstance == nil {
			fmt.Println("Running DB instance now.", viper.GetString("database.file"))

			sqliteDatabase, err := sql.Open("sqlite3", viper.GetString("database.file")) // Open the created SQLite File
			if err != nil {
				panic(fmt.Errorf("fatal error db file: %w", err))
			}

			userService := createUserService(sqliteDatabase)
			presenceService := createPresenceService(sqliteDatabase)

			dbInstance = &DB{
				UserService:     userService,
				PresenceService: presenceService,
				db:              sqliteDatabase,
			}
		}
	}
	return dbInstance
}
