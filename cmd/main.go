package main

import (
	"net/http"

	"github.com/ashX04/new_website/internal/handlers"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Create a secure random key
	key := []byte("your-secure-secret-key-min-32-bytes-long")

	// Initialize the cookie store with additional options
	store := cookie.NewStore(key)
	store.Options(sessions.Options{
		Path:     "/",       // Path for the cookie
		MaxAge:   3600 * 24, // 24 hours
		Secure:   false,     // Set to true in production with HTTPS
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Use sessions middleware
	r.Use(sessions.Sessions("session-name", store))

	// Serve static files
	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads")

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
	r.POST("/login", handlers.LoginProcess)
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
		authorized.GET("/download/:id", handlers.DownloadFile)
		authorized.DELETE("/files/:id", handlers.DeleteFile)
	}

	// Start the server
	r.Run(":8080")
}
