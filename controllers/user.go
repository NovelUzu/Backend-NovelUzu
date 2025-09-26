package controllers

import (
	"NovelUzu/middleware"
	models "NovelUzu/models/postgres"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// @Summary Obtener todos los usuarios
// @Description Retorna una lista de todos los usuarios con su información básica
// @Tags users
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {array} object{username=string,email=string,role=string,status=string,avatar_url=string,created_at=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /user/allusers [get]
func GetAllUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar autenticación
		_, err := middleware.JWT_decoder(c, db)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			return
		}

		var users []models.User

		// Obtener todos los usuarios
		result := db.Find(&users)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener usuarios"})
			return
		}

		// Crear un slice de datos simplificados de usuarios
		simplifiedUsers := make([]gin.H, len(users))
		for i, user := range users {
			userInfo := gin.H{
				"username":   user.ProfileUsername,
				"email":      user.Email,
				"role":       string(user.Role),
				"status":     string(user.Status),
				"created_at": user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}

			// Agregar avatar_url solo si no es nil
			if user.AvatarURL != nil {
				userInfo["avatar_url"] = *user.AvatarURL
			}

			simplifiedUsers[i] = userInfo
		}

		c.JSON(http.StatusOK, simplifiedUsers)
	}
}

// uploadToNextcloud sube un archivo a Nextcloud y retorna la URL pública
func uploadToNextcloud(file multipart.File, filename string) (string, error) {
	// Configuración de Nextcloud
	nextcloudURL := "https://nextcloud.eslus.org/remote.php/dav/files/"
	username := os.Getenv("NEXTCLOUD_USERNAME")
	password := os.Getenv("NEXTCLOUD_PASSWORD")

	if username == "" || password == "" {
		return "", fmt.Errorf("credenciales de Nextcloud no configuradas")
	}

	// Generar nombre único para el archivo
	ext := filepath.Ext(filename)
	uniqueFilename := fmt.Sprintf("avatar_%d%s", time.Now().Unix(), ext)

	// Primero crear el directorio avatars si no existe
	avatarsDir := fmt.Sprintf("%s%s/avatars", nextcloudURL, username)

	// Crear directorio avatars
	req, err := http.NewRequest("MKCOL", avatarsDir, nil)
	if err == nil {
		req.SetBasicAuth(username, password)
		client := &http.Client{Timeout: 10 * time.Second}
		resp, _ := client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
	}

	uploadPath := fmt.Sprintf("avatars/%s", uniqueFilename)
	fullURL := fmt.Sprintf("%s%s/%s", nextcloudURL, username, uploadPath)

	// Leer el contenido del archivo
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("error al leer archivo: %v", err)
	}

	// Crear request HTTP PUT para subir el archivo
	req, err = http.NewRequest("PUT", fullURL, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("error al crear request: %v", err)
	}

	// Configurar autenticación básica
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/octet-stream")

	// Realizar la subida
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error de conexión con Nextcloud: %v", err)
	}
	defer resp.Body.Close()

	// Leer el cuerpo de la respuesta para debug
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error en Nextcloud - Código: %d, URL: %s, Respuesta: %s", resp.StatusCode, fullURL, string(body))
	}

	// Crear enlace público para el archivo
	publicURL, err := crearEnlacePublico(uploadPath, username, password)
	if err != nil {
		// Si falla la creación del enlace público, retornar la URL directa
		fmt.Printf("Error al crear enlace público: %v\n", err)
		directURL := fmt.Sprintf("https://nextcloud.eslus.org/remote.php/dav/files/%s/%s", username, uploadPath)
		return directURL, nil
	}

	return publicURL, nil
}

// crearEnlacePublico crea un enlace público para un archivo en Nextcloud
func crearEnlacePublico(ruta, username, password string) (string, error) {
	nextcloudURL := "https://nextcloud.eslus.org"
	url := nextcloudURL + "/ocs/v2.php/apps/files_sharing/api/v1/shares"
	data := "path=/" + ruta + "&shareType=3&permissions=1"

	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("error al crear request: %v", err)
	}

	req.Header.Set("OCS-APIRequest", "true")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(username, password)

	// Debug: imprimir la URL y datos que se están enviando
	fmt.Printf("Creando enlace público - URL: %s, Data: %s\n", url, data)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error de conexión: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error al leer respuesta: %v", err)
	}

	// Debug: imprimir la respuesta XML completa
	fmt.Printf("Respuesta XML crearEnlacePublico (Status: %d): %s\n", resp.StatusCode, string(body))

	// Trabajar directamente con XML - verificar si el status es ok
	if strings.Contains(string(body), "<status>ok</status>") {
		return extraerTokenDeXML(string(body))
	} else {
		// Extraer el mensaje de error del XML
		messageStart := strings.Index(string(body), "<message>")
		messageEnd := strings.Index(string(body), "</message>")
		if messageStart != -1 && messageEnd != -1 && messageStart < messageEnd {
			errorMsg := string(body)[messageStart+9 : messageEnd]
			return "", fmt.Errorf("error en Nextcloud: %s", errorMsg)
		}
		return "", fmt.Errorf("error en respuesta XML de Nextcloud")
	}
}

// extraerTokenDeXML extrae el token de compartición de una respuesta XML de Nextcloud
func extraerTokenDeXML(xmlResponse string) (string, error) {
	// Primero intentar extraer la URL completa que ya viene formateada
	urlStart := strings.Index(xmlResponse, "<url>")
	urlEnd := strings.Index(xmlResponse, "</url>")

	if urlStart != -1 && urlEnd != -1 && urlStart < urlEnd {
		url := xmlResponse[urlStart+5 : urlEnd]
		return url, nil
	}

	// Si no encuentra URL, buscar el token para construir la URL manualmente
	tokenStart := strings.Index(xmlResponse, "<token>")
	tokenEnd := strings.Index(xmlResponse, "</token>")

	if tokenStart == -1 || tokenEnd == -1 || tokenStart >= tokenEnd {
		return "", fmt.Errorf("no se pudo extraer token ni URL de la respuesta XML")
	}

	token := xmlResponse[tokenStart+7 : tokenEnd]
	publicURL := fmt.Sprintf("https://nextcloud.eslus.org/s/%s", token)
	return publicURL, nil
}

// @Summary Actualizar perfil de usuario
// @Description Actualiza la información del perfil del usuario incluyendo avatar
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param username formData string false "Nuevo nombre de usuario"
// @Param bio formData string false "Biografía del usuario"
// @Param birth_date formData string false "Fecha de nacimiento (YYYY-MM-DD)"
// @Param country formData string false "País del usuario"
// @Param avatar formData file false "Imagen de avatar (JPG, PNG)"
// @Success 200 {object} object{message=string,user=object}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /user/update [put]
func UpdateProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar autenticación
		email, err := middleware.JWT_decoder(c, db)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			return
		}

		// Buscar usuario actual
		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
			return
		}

		// Crear mapa para actualizaciones
		updates := make(map[string]interface{})

		// Procesar campos de texto
		if username := c.PostForm("username"); username != "" {
			// Verificar que el username no esté en uso por otro usuario
			var existingUser models.User
			if err := db.Where("username = ? AND email != ?", username, email).First(&existingUser).Error; err == nil {
				c.JSON(http.StatusConflict, gin.H{"error": "El nombre de usuario ya está en uso"})
				return
			}
			updates["username"] = username
		}

		if bio := c.PostForm("bio"); bio != "" {
			updates["bio"] = bio
		}

		if birthDate := c.PostForm("birth_date"); birthDate != "" {
			parsedDate, err := time.Parse("2006-01-02", birthDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha inválido. Use YYYY-MM-DD"})
				return
			}
			updates["birth_date"] = parsedDate
		}

		if country := c.PostForm("country"); country != "" {
			updates["country"] = country
		}

		// Procesar archivo de avatar
		file, header, err := c.Request.FormFile("avatar")
		if err == nil {
			defer file.Close()

			// Validar tipo de archivo
			contentType := header.Header.Get("Content-Type")
			if !strings.HasPrefix(contentType, "image/") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "El archivo debe ser una imagen"})
				return
			}

			// Validar tamaño (max 5MB)
			if header.Size > 5*1024*1024 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "El archivo es demasiado grande (máximo 5MB)"})
				return
			}

			// Intentar subir a Nextcloud
			avatarURL, err := uploadToNextcloud(file, header.Filename)
			if err != nil {
				// Si falla la subida a Nextcloud, log el error pero continúa sin la imagen
				fmt.Printf("Error al subir a Nextcloud: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al subir imagen. Verifique la configuración de Nextcloud"})
				return
			}

			updates["avatar_url"] = avatarURL
		}

		// Si no hay actualizaciones, retornar error
		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No se proporcionaron campos para actualizar"})
			return
		}

		// Actualizar timestamp
		updates["updated_at"] = time.Now()

		// Realizar actualización en la base de datos
		if err := db.Model(&user).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar perfil"})
			return
		}

		// Buscar usuario actualizado para retornar
		var updatedUser models.User
		if err := db.Where("email = ?", email).First(&updatedUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener usuario actualizado"})
			return
		}

		// Preparar respuesta
		userResponse := gin.H{
			"email":          updatedUser.Email,
			"username":       updatedUser.ProfileUsername,
			"role":           string(updatedUser.Role),
			"status":         string(updatedUser.Status),
			"email_verified": updatedUser.EmailVerified,
			"created_at":     updatedUser.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":     updatedUser.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Agregar campos opcionales
		if updatedUser.AvatarURL != nil {
			userResponse["avatar_url"] = *updatedUser.AvatarURL
		}
		if updatedUser.Bio != nil {
			userResponse["bio"] = *updatedUser.Bio
		}
		if updatedUser.BirthDate != nil {
			userResponse["birth_date"] = updatedUser.BirthDate.Format("2006-01-02")
		}
		if updatedUser.Country != nil {
			userResponse["country"] = *updatedUser.Country
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Perfil actualizado exitosamente",
			"user":    userResponse,
		})
	}
}

// @Summary Cambiar contraseña del usuario
// @Description Permite cambiar la contraseña del usuario después de verificar la contraseña actual
// @Tags users
// @Accept x-www-form-urlencoded
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param current_password formData string true "Contraseña actual"
// @Param new_password formData string true "Nueva contraseña"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 403 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /user/change-password [put]
func ChangePassword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar autenticación
		email, err := middleware.JWT_decoder(c, db)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			return
		}

		// Obtener parámetros
		currentPassword := c.PostForm("current_password")
		newPassword := c.PostForm("new_password")

		// Validar que se proporcionaron ambos campos
		if currentPassword == "" || newPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "current_password y new_password son obligatorios"})
			return
		}

		// Validar longitud de nueva contraseña
		if len(newPassword) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "La nueva contraseña debe tener al menos 6 caracteres"})
			return
		}

		// Buscar usuario actual
		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
			return
		}

		// Verificar contraseña actual
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "La contraseña actual es incorrecta"})
			return
		}

		// Verificar que la nueva contraseña sea diferente a la actual
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(newPassword)); err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "La nueva contraseña debe ser diferente a la actual"})
			return
		}

		// Hashear nueva contraseña
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar la nueva contraseña"})
			return
		}

		// Actualizar contraseña en la base de datos
		if err := db.Model(&user).Updates(map[string]interface{}{
			"password_hash": string(hashedPassword),
			"updated_at":    time.Now(),
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar la contraseña"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Contraseña cambiada exitosamente",
		})
	}
}

// @Summary Eliminar cuenta de usuario
// @Description Elimina permanentemente la cuenta del usuario después de verificar la contraseña
// @Tags users
// @Accept x-www-form-urlencoded
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param password formData string true "Contraseña actual para confirmar eliminación"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 403 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /user/delete-account [delete]
func DeleteAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar autenticación
		email, err := middleware.JWT_decoder(c, db)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			return
		}

		// Obtener contraseña de confirmación
		password := c.PostForm("password")
		if password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "La contraseña es obligatoria para confirmar la eliminación"})
			return
		}

		// Buscar usuario actual
		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
			return
		}

		// Verificar contraseña actual
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Contraseña incorrecta"})
			return
		}

		// Eliminar el usuario de la base de datos
		if err := db.Delete(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar la cuenta"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Cuenta eliminada exitosamente",
		})
	}
}
