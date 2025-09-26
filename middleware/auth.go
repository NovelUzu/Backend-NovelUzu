package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"NovelUzu/constants/auth"

	models "NovelUzu/models/postgres"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// AuthRequired is a simple middleware to check the session.
func AuthRequired(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		c.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		c.Abort()
		return
	}

	secret := os.Getenv("KEY")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificar el método de firma
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unauthorized")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		c.Abort()
		return
	}
	c.Next()
}

func JWT_decoder(c *gin.Context, db *gorm.DB) (string, error) {
	// Obtener el token del encabezado Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("unauthorized")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", fmt.Errorf("unauthorized")
	}

	// Parsear el JWT
	secret := os.Getenv("KEY")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificar el método de firma
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unauthorized")
		}
		return []byte(secret), nil
	})

	// Verificar si el token es válido
	if err != nil || !token.Valid {
		return "", fmt.Errorf("unauthorized")
	}

	// Obtener los datos del token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("unauthorized")
	}

	// Verificar que el email existe en los claims
	emailClaim, exists := claims[auth.Email]
	if !exists {
		return "", fmt.Errorf("unauthorized")
	}

	email, ok := emailClaim.(string)
	if !ok {
		return "", fmt.Errorf("unauthorized")
	}

	// Verificar que el usuario existe en la base de datos (solo si db no es nil)
	if db != nil {
		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			return "", fmt.Errorf("unauthorized")
		}
	}

	return email, nil
}

func Socketio_JWT_decoder(authData map[string]interface{}) (string, error) {
	// Obtener el token del authData
	tokenStringRaw, ok := authData["authorization"].(string)
	if !ok {
		return "", fmt.Errorf("unauthorized")
	}

	tokenString := strings.TrimPrefix(tokenStringRaw, "Bearer ")

	// Parsear el JWT
	secret := os.Getenv("KEY")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificar el método de firma
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unauthorized")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("unauthorized")
	}

	// Obtener los datos del token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("unauthorized")
	}

	// Verificar que el email existe en los claims
	emailClaim, exists := claims[auth.Email]
	if !exists {
		return "", fmt.Errorf("unauthorized")
	}

	email, ok := emailClaim.(string)
	if !ok {
		return "", fmt.Errorf("unauthorized")
	}

	return email, nil
}

// me is the handler that will return the user information stored in the
// session.
func me(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(auth.Email)
	c.JSON(http.StatusOK, gin.H{"usuario": user})
}

// status is the handler that will tell the user whether it is logged in or not.
func status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"estado": "Sesión activa"})
}
