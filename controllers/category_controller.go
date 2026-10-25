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

const categoryColumns = "id, name, created_at, created_by, modified_at, modified_by"

func scanCategory(row interface{ Scan(...any) error }, k *models.Category) error {
	return row.Scan(&k.ID, &k.Name, &k.CreatedAt, &k.CreatedBy, &k.ModifiedAt, &k.ModifiedBy)
}

func categoryNotFound(c *gin.Context, id int) {
	c.JSON(http.StatusNotFound, gin.H{"error": "kategori dengan ID " + strconv.Itoa(id) + " tidak ditemukan"})
}

// bindCategory membaca body JSON dan memvalidasi field wajib
func bindCategory(c *gin.Context) (models.CategoryInput, bool) {
	var input models.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data JSON tidak valid: " + err.Error()})
		return input, false
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name tidak boleh kosong"})
		return input, false
	}
	return input, true
}

// GET /api/categories
func GetAllCategories(c *gin.Context) {
	rows, err := config.DB.Query("SELECT " + categoryColumns + " FROM categories ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data: " + err.Error()})
		return
	}
	defer rows.Close()

	categories := []models.Category{}
	for rows.Next() {
		var k models.Category
		if err := scanCategory(rows, &k); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membaca data: " + err.Error()})
			return
		}
		categories = append(categories, k)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membaca data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// POST /api/categories
func CreateCategory(c *gin.Context) {
	input, ok := bindCategory(c)
	if !ok {
		return
	}

	var k models.Category
	err := scanCategory(config.DB.QueryRow(
		"INSERT INTO categories (name, created_by) VALUES ($1, $2) RETURNING "+categoryColumns,
		input.Name, currentUser(c),
	), &k)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan data: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "kategori berhasil ditambahkan",
		"data":    k,
	})
}

// GET /api/categories/:id
func GetCategoryByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var k models.Category
	err := scanCategory(config.DB.QueryRow("SELECT "+categoryColumns+" FROM categories WHERE id = $1", id), &k)
	if errors.Is(err, sql.ErrNoRows) {
		categoryNotFound(c, id)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": k})
}

// PUT /api/categories/:id
func UpdateCategory(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	input, ok := bindCategory(c)
	if !ok {
		return
	}

	var k models.Category
	err := scanCategory(config.DB.QueryRow(
		"UPDATE categories SET name = $1, modified_at = NOW(), modified_by = $2 WHERE id = $3 RETURNING "+categoryColumns,
		input.Name, currentUser(c), id,
	), &k)
	if errors.Is(err, sql.ErrNoRows) {
		categoryNotFound(c, id)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memperbarui data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "kategori berhasil diperbarui",
		"data":    k,
	})
}

// DELETE /api/categories/:id
func DeleteCategory(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	result, err := config.DB.Exec("DELETE FROM categories WHERE id = $1", id)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23503" {
		c.JSON(http.StatusConflict, gin.H{"error": "kategori masih digunakan oleh buku, hapus bukunya terlebih dahulu"})
		return
	}
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
		categoryNotFound(c, id)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "kategori berhasil dihapus"})
}

// GET /api/categories/:id/books
func GetBooksByCategory(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var exists bool
	if err := config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)", id).Scan(&exists); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data: " + err.Error()})
		return
	}
	if !exists {
		categoryNotFound(c, id)
		return
	}

	rows, err := config.DB.Query("SELECT "+bookColumns+" FROM books WHERE category_id = $1 ORDER BY id", id)
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
