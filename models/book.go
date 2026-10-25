package models

import "time"

type Book struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ImageURL    string     `json:"image_url"`
	ReleaseYear int        `json:"release_year"`
	Price       int        `json:"price"`
	TotalPage   int        `json:"total_page"`
	Thickness   string     `json:"thickness"`
	CategoryID  int        `json:"category_id"`
	CreatedAt   time.Time  `json:"created_at"`
	CreatedBy   *string    `json:"created_by"`
	ModifiedAt  *time.Time `json:"modified_at"`
	ModifiedBy  *string    `json:"modified_by"`
}

// BookInput adalah body JSON dari client, thickness tidak diinput tapi dihitung dari total_page
type BookInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	ReleaseYear int    `json:"release_year"`
	Price       int    `json:"price"`
	TotalPage   int    `json:"total_page"`
	CategoryID  int    `json:"category_id"`
}

// ThicknessFromPages mengonversi total halaman menjadi ketebalan buku
func ThicknessFromPages(totalPage int) string {
	if totalPage > 100 {
		return "tebal"
	}
	return "tipis"
}
