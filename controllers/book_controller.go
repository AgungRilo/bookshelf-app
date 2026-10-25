package controllers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/AgungRilo/bookshelf-app/config"
	"github.com/AgungRilo/bookshelf-app/models"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

const (
	bookColumns    = "id, title, description, image_url, release_year, price, total_page, thickness, category_id, created_at, created_by, modified_at, modified_by"
	minReleaseYear = 1980
	maxReleaseYear = 2024
)

func scanBook(row interface{ Scan(...any) error }, b *models.Book) error {
	return row.Scan(&b.ID, &b.Title, &b.Description, &b.ImageURL, &b.ReleaseYear, &b.Price,
		&b.TotalPage, &b.Thickness, &b.CategoryID, &b.CreatedAt, &b.CreatedBy, &b.ModifiedAt, &b.ModifiedBy)
}

func collectBooks(rows *sql.Rows) ([]models.Book, error) {
	books := []models.Book{}
	for rows.Next() {
		var b models.Book
		if err := scanBook(rows, &b); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, rows.Err()
}

func bookNotFound(c *gin.Context, id int) {
	c.JSON(http.StatusNotFound, gin.H{"error": "buku dengan ID " + strconv.Itoa(id) + " tidak ditemukan"})
}

// isInvalidCategory mengecek apakah error berasal dari foreign key category_id
func isInvalidCategory(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}

// bindBook membaca body JSON dan memvalidasi field wajib
func bindBook(c *gin.Context) (models.BookInput, bool) {
	var input models.BookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data JSON tidak valid: " + err.Error()})
		return input, false
	}

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title tidak boleh kosong"})
		return input, false
	}
	if input.ReleaseYear < minReleaseYear || input.ReleaseYear > maxReleaseYear {
		c.JSON(http.StatusBadRequest, gin.H{"error": "release_year harus di antara 1980 sampai 2024"})
		return input, false
	}
	if input.Price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price tidak boleh negatif"})
		return input, false
	}
	if input.TotalPage <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "total_page harus lebih dari 0"})
		return input, false
	}
	if input.CategoryID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category_id wajib diisi"})
		return input, false
	}
	return input, true
}

// GET /api/books
func GetAllBooks(c *gin.Context) {
	rows, err := config.DB.Query("SELECT " + bookColumns + " FROM books ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data: " + err.Error()})
		return
	}
	defer rows.Close()

	books, err := collectBooks(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membaca data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": books})
}

// POST /api/books
func CreateBook(c *gin.Context) {
	input, ok := bindBook(c)
	if !ok {
		return
	}

	var b models.Book
	err := scanBook(config.DB.QueryRow(
		`INSERT INTO books (title, description, image_url, release_year, price, total_page, thickness, category_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING `+bookColumns,
		input.Title, input.Description, input.ImageURL, input.ReleaseYear, input.Price,
		input.TotalPage, models.ThicknessFromPages(input.TotalPage), input.CategoryID, currentUser(c),
	), &b)
	if isInvalidCategory(err) {
		categoryNotFound(c, input.CategoryID)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan data: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "buku berhasil ditambahkan",
		"data":    b,
	})
}

// GET /api/books/:id
func GetBookByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var b models.Book
	err := scanBook(config.DB.QueryRow("SELECT "+bookColumns+" FROM books WHERE id = $1", id), &b)
	if errors.Is(err, sql.ErrNoRows) {
		bookNotFound(c, id)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": b})
}

// PUT /api/books/:id
func UpdateBook(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	input, ok := bindBook(c)
	if !ok {
		return
	}

	var b models.Book
	err := scanBook(config.DB.QueryRow(
		`UPDATE books SET title = $1, description = $2, image_url = $3, release_year = $4, price = $5,
		total_page = $6, thickness = $7, category_id = $8, modified_at = NOW(), modified_by = $9
		WHERE id = $10 RETURNING `+bookColumns,
		input.Title, input.Description, input.ImageURL, input.ReleaseYear, input.Price,
		input.TotalPage, models.ThicknessFromPages(input.TotalPage), input.CategoryID, currentUser(c), id,
	), &b)
	if errors.Is(err, sql.ErrNoRows) {
		bookNotFound(c, id)
		return
	}
	if isInvalidCategory(err) {
		categoryNotFound(c, input.CategoryID)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memperbarui data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "buku berhasil diperbarui",
		"data":    b,
	})
}

// DELETE /api/books/:id
func DeleteBook(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	result, err := config.DB.Exec("DELETE FROM books WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menghapus data: " + err.Error()})
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menghapus data: " + err.Error()})
		return
	}
	if affected == 0 {
		bookNotFound(c, id)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "buku berhasil dihapus"})
}
