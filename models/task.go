package models 

import (
	"context"
	"time"

	"taskapi/db"
)

//Task mirrors the tasks table in PostgreSQL
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Done		bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
}

//CreateTable runs on startup - creates the table if it doesn't exist
func CreateTable() error {
	query := `
CREATE TABLE IF NOT EXISTS tasks (
	id SERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	done BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`
	_, err := db.DB.Exec(context.Background(), query)
	return err
}