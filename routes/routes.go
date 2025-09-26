package routes

import (
	"NovelUzu/controllers"
	"NovelUzu/middleware"
	utils "NovelUzu/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, db *gorm.DB) {

	// utils global
	router.Use(utils.ErrorHandler())

	// Swagger route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Testing a basic endpoint, and the auto-docs

	// API routes group
	api := router.Group("/")
	api.POST("/login", controllers.Login(db))
	api.POST("/signup", controllers.SignUp(db))

	// Rutas autenticadas
	auth := api.Group("/auth")
	auth.Use(middleware.AuthRequired)
	{
		auth.DELETE("/logout", controllers.Logout)
		auth.GET("/verify-token", controllers.VerifyTokenAndGetUser(db))
	}

	user := api.Group("/user")
	user.Use(middleware.AuthRequired)
	{
		user.GET("/allusers", controllers.GetAllUsers(db))
		user.PUT("/update", controllers.UpdateProfile(db))
		user.PUT("/change-password", controllers.ChangePassword(db))
		user.DELETE("/delete-account", controllers.DeleteAccount(db))
	}
}
