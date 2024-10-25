package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ShowDashboard(c *gin.Context) {
	// Render the dashboard template
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Dashboard",
	})
}
