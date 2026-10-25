package models

import "time"

type Category struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	CreatedBy  *string    `json:"created_by"`
	ModifiedAt *time.Time `json:"modified_at"`
	ModifiedBy *string    `json:"modified_by"`
}

type CategoryInput struct {
	Name string `json:"name"`
}
