package main

import (
	"encoding/json"
	"os"
)

type Database struct {
	Boards []*Board `json:"boards"`
}

type Board struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Status bool    `json:"status"`
	Tasks  []*Task `json:"tasks"`
}

type Task struct {
	ID     int64  `json:"id"`
	Text   string `json:"name"`
	Status bool   `json:"status"`
}

func NewDatabase() *Database {
	return &Database{[]*Board{}}
}

func ReadDatabaseFromFile(name string) (*Database, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var db Database
	err = json.NewDecoder(file).Decode(&db)
	return &db, err
}

func (db *Database) WriteToFile(name string) error {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(db)
}
