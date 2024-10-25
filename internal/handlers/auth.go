package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/sessions"
)

// PocketBase API URL
const baseURL = "http://localhost:8090/api/collections/users/records"

// Session store
var store = sessions.NewCookieStore([]byte("your-secret-key"))

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
func LoginProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	// Prepare login request data
	loginData := map[string]interface{}{
		"identity": email,
		"password": password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		return
	}

	// Send POST request to PocketBase API for authentication
	resp, err := http.Post("http://localhost:8090/api/collections/users/auth-with-password", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Failed to log in", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var authResponse AuthResponse
		if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
			http.Error(w, "Failed to parse response", http.StatusInternalServerError)
			return
		}

		// Store the user ID in the session
		session, _ := store.Get(r, "session-name")
		session.Values["userID"] = authResponse.Record.Id
		session.Save(r, w)

		fmt.Println("User logged in successfully.")
		http.Redirect(w, r, "/?message=Login successful!", http.StatusSeeOther)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("Error logging in: %s", string(body)), http.StatusInternalServerError)
	}
}

// AuthResponse structure for decoding login response
type AuthResponse struct {
	Record struct {
		Id string `json:"id"`
	} `json:"record"`
}
