// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package pgdb

import (
	"time"
)

type Setting struct {
	ID                    string
	NotionToken           string
	NotionPageID          string
	MeuPluggyClientID     string
	MeuPluggyClientSecret string
	MeuPluggyAccountIds   []string
	UserID                string
	UpdatedAt             time.Time
}

type User struct {
	ID        string
	Email     string
	UpdatedAt time.Time
}
