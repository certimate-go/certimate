package domain

import (
	"time"
)

type Meta struct {
	Id        string    `db:"id"      json:"id"`
	CreatedAt time.Time `db:"created" json:"created"`
	UpdatedAt time.Time `db:"updated" json:"updated"`
}
