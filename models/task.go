package models

import (
	"context"
	"time"

	"taskapi/db"
)

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS tasks (
			id         SERIAL PRIMARY KEY,
			title      TEXT        NOT NULL,
			done       BOOLEAN     NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`
	_, err := db.DB.Exec(context.Background(), query)
	return err
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

func CreateUsersTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id       SERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);
	`
	_, err := db.DB.Exec(context.Background(), query)
	return err
}