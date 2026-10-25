package models

import "time"

type User struct {
	ID         int        `json:"id"`
	Username   string     `json:"username"`
	Password   string     `json:"-"`
	CreatedAt  time.Time  `json:"created_at"`
	CreatedBy  *string    `json:"created_by"`
	ModifiedAt *time.Time `json:"modified_at"`
	ModifiedBy *string    `json:"modified_by"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
