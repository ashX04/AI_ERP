package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ShowHome(c *gin.Context) {
	// Check if user is authenticated
	session, _ := store.Get(c.Request, "session-name")
	isAuthenticated := session.Values["userID"] != nil

	c.HTML(http.StatusOK, "home.html", gin.H{
		"title":           "Home",
		"IsAuthenticated": isAuthenticated,
	})
}

// Logout handler
func Logout(c *gin.Context) {
	session, _ := store.Get(c.Request, "session-name")
	session.Values["userID"] = nil
	session.Save(c.Request, c.Writer)
	c.Redirect(http.StatusSeeOther, "/")
}
