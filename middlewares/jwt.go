package middlewares

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AgungRilo/bookshelf-app/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	return []byte(config.GetEnv("JWT_SECRET", "rahasia-bookshelf"))
}

// GenerateToken membuat token JWT yang berlaku 24 jam
func GenerateToken(userID int, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret())
}

// JWTAuth memastikan request membawa header "Authorization: Bearer <token>" yang valid
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenString, found := strings.CutPrefix(header, "Bearer ")
		if !found || tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token tidak ditemukan, gunakan header Authorization: Bearer <token>"})
			return
		}

		claims := &Claims{}
		_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("metode signing tidak valid")
			}
			return jwtSecret(), nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token tidak valid atau sudah kedaluwarsa"})
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}
