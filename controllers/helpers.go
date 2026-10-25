package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseID membaca parameter :id dan memastikan berupa angka positif
func parseID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID harus berupa angka positif"})
		return 0, false
	}
	return id, true
}

// currentUser mengambil username dari token JWT yang sudah divalidasi middleware
func currentUser(c *gin.Context) string {
	return c.GetString("username")
}
