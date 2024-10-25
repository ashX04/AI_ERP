package main

import (
	"net/http"

	"github.com/ashX04/new_website/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")

	// Load HTML templates
	r.LoadHTMLGlob("internal/templates/*")

	// Public routes
	r.GET("/", handlers.ShowHome)
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
	r.POST("/register", gin.WrapF(handlers.RegisterProcess))
	r.POST("/login", gin.WrapF(handlers.LoginProcess))
	r.GET("/logout", handlers.Logout)

	// Protected routes (require authentication)
	authorized := r.Group("/")
	authorized.Use(handlers.RequireAuth())
	{
		authorized.GET("/dashboard", handlers.ShowDashboard)
		authorized.GET("/upload", func(c *gin.Context) {
			c.HTML(http.StatusOK, "upload.html", nil)
		})
		authorized.POST("/upload", handlers.UploadImage)
		authorized.GET("/download/:fileID", handlers.DownloadFile)
	}

	// Start the server
	r.Run(":8080")
}
