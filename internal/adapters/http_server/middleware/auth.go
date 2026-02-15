package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const UserIDKey = "userID"

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		userID, err := parseJWTToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		userIDParsed, err := uuid.Parse(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		c.Set(UserIDKey, userIDParsed)
		c.Next()
	}
}

func parseJWTToken(tokenString string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return "", err
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	return userID, nil
}

func GetUserIDFromContext(c *gin.Context) (string, error) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", errors.New("userID not found in context")
	}

	return userID.(string), nil
}
