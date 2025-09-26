package controllers

import (
	"NovelUzu/constants/auth"
	"NovelUzu/middleware"
	models "NovelUzu/models/postgres"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// @Summary Iniciar sesión
// @Description Autentica un usuario y crea una sesión
// @Tags auth
// @Accept x-www-form-urlencoded
// @Produce json
// @Param email formData string true "Correo electrónico del usuario"
// @Param password formData string true "Contraseña del usuario"
// @Success 200 {object} object{message=string,token=string,user=object{email=string,username=string,role=string,status=string,avatar_url=string,bio=string,birth_date=string,country=string,email_verified=boolean,last_login=string,created_at=string,updated_at=string}}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /login [post]
func Login(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")

		//Minimum input sanitizing
		if strings.Trim(email, " ") == "" || strings.Trim(password, " ") == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Los parametros email y password son obligatorios"})
			return
		}

		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario o contraseña inválidos"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario o contraseña inválidos"})
			return
		}

		// Generate JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			auth.Email: user.Email,
		})

		secret := os.Getenv("KEY")
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar el JWT"})
			return
		}

		// Actualizar last_login
		now := time.Now()
		user.LastLogin = &now
		db.Save(&user)

		// Preparar la información del usuario
		userInfo := gin.H{
			"email":          user.Email,
			"username":       user.ProfileUsername,
			"role":           string(user.Role),
			"status":         string(user.Status),
			"email_verified": user.EmailVerified,
			"created_at":     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Agregar campos opcionales solo si no son nil
		if user.AvatarURL != nil {
			userInfo["avatar_url"] = *user.AvatarURL
		}
		if user.Bio != nil {
			userInfo["bio"] = *user.Bio
		}
		if user.BirthDate != nil {
			userInfo["birth_date"] = user.BirthDate.Format("2006-01-02")
		}
		if user.Country != nil {
			userInfo["country"] = *user.Country
		}
		if user.LastLogin != nil {
			userInfo["last_login"] = user.LastLogin.Format("2006-01-02T15:04:05Z07:00")
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Inicio de sesión exitoso.",
			"token":   tokenString,
			"user":    userInfo,
		})
	}
}

// @Summary Cerrar sesión
// @Description Termina la sesión del usuario
// @Tags auth
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /auth/logout [delete]
func Logout(c *gin.Context) {
	//This serves no purpose with JWT so TODO rething
	_, err := middleware.JWT_decoder(c, nil)
	if err != nil {
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"mensaje": "Cierre de sesión exitoso"})
}

// @Summary Verificar token JWT
// @Description Verifica si un token JWT es válido y devuelve información del usuario
// @Tags auth
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} object{email=string,username=string,role=string,status=string,avatar_url=string,bio=string,birth_date=string,country=string,email_verified=boolean,last_login=string,created_at=string,updated_at=string}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /auth/verify-token [get]
func VerifyTokenAndGetUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extraer el email del token JWT
		email, err := middleware.JWT_decoder(c, db)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			return
		}

		// Buscar el usuario en la base de datos
		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar el usuario"})
			return
		}

		// Actualizar last_login
		now := time.Now()
		user.LastLogin = &now
		db.Save(&user)

		// Preparar la respuesta con la información del usuario
		userInfo := gin.H{
			"email":          user.Email,
			"username":       user.ProfileUsername,
			"role":           string(user.Role),
			"status":         string(user.Status),
			"email_verified": user.EmailVerified,
			"created_at":     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Agregar campos opcionales solo si no son nil
		if user.AvatarURL != nil {
			userInfo["avatar_url"] = *user.AvatarURL
		}
		if user.Bio != nil {
			userInfo["bio"] = *user.Bio
		}
		if user.BirthDate != nil {
			userInfo["birth_date"] = user.BirthDate.Format("2006-01-02")
		}
		if user.Country != nil {
			userInfo["country"] = *user.Country
		}
		if user.LastLogin != nil {
			userInfo["last_login"] = user.LastLogin.Format("2006-01-02T15:04:05Z07:00")
		}

		c.JSON(http.StatusOK, userInfo)
	}
}

// @Summary Registrar nuevo usuario
// @Description Crea una nueva cuenta de usuario
// @Tags auth
// @Accept x-www-form-urlencoded
// @Produce json
// @Param username formData string true "Nombre de usuario"
// @Param email formData string true "Correo electrónico"
// @Param password formData string true "Contraseña"
// @Success 201 {object} object{message=string,user=object{username=string,email=string}}
// @Failure 400 {object} object{error=string}
// @Failure 409 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /signup [post]
func SignUp(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		email := c.PostForm("email")
		password := c.PostForm("password")

		// Minimum input sanitizing
		if strings.TrimSpace(username) == "" || strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Los parametros username, email y password son obligatorios"})
			return
		}

		// Check if user already exists
		var existingUser models.User
		if err := db.Where("email = ? OR username = ?", email, username).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "El usuario o email ya existe"})
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al hashear la contraseña"})
			return
		}
		// Create User
		user := models.User{
			Email:           email,
			ProfileUsername: username,
			PasswordHash:    string(hashedPassword),
			CreatedAt:       time.Now(),
		}

		if err := db.Create(&user).Error; err != nil {
			// Rollback game profile creation if user creation fails
			// TODO: do it with a transaction?
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear el usuario"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Usuario creado exitosamente",
			"user": gin.H{
				"username": username,
				"email":    email,
			},
		})
	}
}
