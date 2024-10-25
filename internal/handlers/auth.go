package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	sessions "github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// PocketBase API URL
const baseURL = "http://localhost:8090/api/collections/users/records"

// Session store
var store = cookie.NewStore([]byte("your-secret-key"))

// Register Process
func RegisterProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	// Prepare data for PocketBase API call
	data := map[string]interface{}{
		"email":           email,
		"password":        password,
		"passwordConfirm": password, // PocketBase requires password confirmation
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		return
	}

	// Send POST request to PocketBase API to register a new user
	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Failed to register", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("Error registering: %s", string(body)), http.StatusInternalServerError)
		return
	}

	fmt.Println("User registered with email:", email)

	// Redirect to login page after registration
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Login Process
func LoginProcess(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
		return
	}

	email := c.PostForm("email")
	password := c.PostForm("password")

	// Prepare login request data
	loginData := map[string]interface{}{
		"identity": email,
		"password": password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request"})
		return
	}

	// Send POST request to PocketBase API for authentication
	resp, err := http.Post("http://localhost:8090/api/collections/users/auth-with-password", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log in"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result struct {
			Token  string `json:"token"`
			Record struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			} `json:"record"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}

		// Store both user ID and token in session
		session := sessions.Default(c)
		session.Set("userID", result.Record.ID)
		session.Set("token", result.Token)
		session.Set("authenticated", true) // Add this explicit authentication flag
		err := session.Save()              // Make sure to check the error
		if err != nil {
			// Handle error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}

		c.Redirect(http.StatusSeeOther, "/dashboard")
		return
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error logging in: %s", string(body))})
	}
}

// AuthResponse structure for decoding login response
type AuthResponse struct {
	Record struct {
		Id string `json:"id"`
	} `json:"record"`
}

// RequireAuth middleware for authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		// Debug: Print session data
		fmt.Printf("Session data: %+v\n", session.Get("authenticated"))

		// Check if user is authenticated
		auth := session.Get("authenticated")
		if auth == nil || auth != true {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// User is authenticated, continue
		c.Next()
	}
}
