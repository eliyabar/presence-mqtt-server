package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

func TestAddPresence(t *testing.T) {

	sqliteDatabase, err := sql.Open("sqlite3", "../mq.db") // Open the created SQLite File
	if err != nil {
		t.Fatalf(`sqliteDatabase ERR ,%v, want "", error`, err)
	}

	defer sqliteDatabase.Close() // Defer Closing the database

	presenceService := createPresenceService(sqliteDatabase)
	err = presenceService.UpsertPresence(1, true)

	if err != nil {
		t.Fatalf(`UpsertPresence ERR ,%v, want "", error`, err)

	}
}

func TestAddPresenceFULLT(t *testing.T) {

	sqliteDatabase, err := sql.Open("sqlite3", `mq.db`) // Open the created SQLite File
	if err != nil {
		t.Fatalf(`sqliteDatabase ERR ,%v, want "", error`, err)
	}

	defer sqliteDatabase.Close() // Defer Closing the database

	presenceService := createPresenceService(sqliteDatabase)
	p, err := presenceService.GetPresence()
	fmt.Printf("presence : %v \n\n", p)
	if err != nil {
		t.Fatalf(`UpsertPresence ERR ,%v, want "", error`, err)

	}
}

type JsonMessage struct {
	Id           string `json:"id"`
	Confidence   string `json:"confidence"`
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	Type         string `json:"type"`
	Retained     string `json:"retained"`
	Timestamp    string `json:"timestamp"`
	Version      string `json:"version"`
}

func TestUnmarshal(t *testing.T) {

	var jsonMessage JsonMessage
	var msg = `{"id":"00:00:00:00:00:00","confidence":"0","name":"TestName","manufacturer":"Phone","type":"KNOWN_MAC","retained":"false","timestamp":"Fri Jun 24 2022 00:32:18 GMT+0100 (BST)","version":"0.2.200"}`
	err := json.Unmarshal([]byte(msg), &jsonMessage)

	if err != nil {
		t.Fatalf("could not Unmarshal %s \n", err)

	}
}
