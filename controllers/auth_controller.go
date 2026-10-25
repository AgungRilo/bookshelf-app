package controllers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/AgungRilo/bookshelf-app/config"
	"github.com/AgungRilo/bookshelf-app/middlewares"
	"github.com/AgungRilo/bookshelf-app/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// POST /api/users/login
func Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data JSON tidak valid: " + err.Error()})
		return
	}
	input.Username = strings.TrimSpace(input.Username)
	if input.Username == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username dan password wajib diisi"})
		return
	}

	var user models.User
	err := config.DB.QueryRow(
		"SELECT id, username, password FROM users WHERE username = $1", input.Username,
	).Scan(&user.ID, &user.Username, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username atau password salah"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data user: " + err.Error()})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username atau password salah"})
		return
	}

	token, err := middlewares.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login berhasil",
		"token":   token,
	})
}
