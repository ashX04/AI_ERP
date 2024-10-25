package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	sessions "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ExcelFile struct {
	ID        string `json:"id"`
	User      string `json:"user"`
	ExcelPath string `json:"excel"`
	ImagePath string `json:"source_image"` // Changed from SourceImage to ImagePath
	Created   string `json:"created"`
	Updated   string `json:"updated"`
	FileName  string `json:"-"`
}

type PocketBaseResponse struct {
	Page       int         `json:"page"`
	PerPage    int         `json:"perPage"`
	TotalItems int         `json:"totalItems"`
	Items      []ExcelFile `json:"items"`
}

func ShowDashboard(c *gin.Context) {
	session := sessions.Default(c)

	// Debug: Print session data
	fmt.Printf("Dashboard session data: %+v\n", session.Get("authenticated"))

	userID := session.Get("userID")
	if userID == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Add debug logging
	log.Printf("User ID from session: %v", userID)

	// Fetch excel files from PocketBase
	resp, err := http.Get("http://localhost:8090/api/collections/excel_files/records")
	if err != nil {
		log.Printf("Error fetching from PocketBase: %v", err)
		c.HTML(http.StatusInternalServerError, "dashboard.html", gin.H{
			"title": "Dashboard",
			"error": "Failed to fetch files",
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		c.HTML(http.StatusInternalServerError, "dashboard.html", gin.H{
			"title": "Dashboard",
			"error": "Failed to read response",
		})
		return
	}

	var pbResp PocketBaseResponse
	if err := json.Unmarshal(body, &pbResp); err != nil {
		log.Printf("Error unmarshaling response: %v", err)
		c.HTML(http.StatusInternalServerError, "dashboard.html", gin.H{
			"title": "Dashboard",
			"error": "Failed to parse response",
		})
		return
	}

	// Filter files for the current user
	var userFiles []ExcelFile
	for _, file := range pbResp.Items {
		if file.User == userID.(string) {
			// Format the file paths to be accessible
			file.ExcelPath = "/uploads/" + file.ExcelPath
			file.ImagePath = "/uploads/" + file.ImagePath // Updated from SourceImage to ImagePath
			// Extract filename from the Excel path
			file.FileName = file.ExcelPath[len("/uploads/"):]
			userFiles = append(userFiles, file)
		}
	}

	// Render the dashboard template with the files
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":  "Dashboard",
		"files":  userFiles,
		"userID": userID,
	})
}

// Helper function to delete a file
func DeleteFile(c *gin.Context) {
	fileID := c.Param("id")

	// Delete from PocketBase
	req, err := http.NewRequest("DELETE", "http://localhost:8090/api/collections/excel_files/records/"+fileID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create delete request"})
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file from database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// Add this new handler function
func DownloadFile(c *gin.Context) {
	fileID := c.Param("id")
	log.Printf("Attempting to download file with ID: %s", fileID)

	// Fetch file details from PocketBase
	resp, err := http.Get("http://localhost:8090/api/collections/excel_files/records/" + fileID)
	if err != nil {
		log.Printf("Error fetching file details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch file details"})
		return
	}
	defer resp.Body.Close()

	var file ExcelFile
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		log.Printf("Error parsing file details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse file details"})
		return
	}

	// Construct the file path using PocketBase storage location
	pbStoragePath := "D:\\codes\\golang\\basedatabase\\pb_data\\storage\\raon6fxd96tluwr"
	filePath := filepath.Join(pbStoragePath, fileID, file.ExcelPath)
	log.Printf("Constructed file path: %s", filePath)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("File not found: %s", filePath)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Open the file
	fileContent, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer fileContent.Close()

	// Set headers for file download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(file.ExcelPath)))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	// Serve the file to the client
	log.Printf("Serving file: %s", filePath)
	if _, err := io.Copy(c.Writer, fileContent); err != nil {
		log.Printf("Error serving file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serve file"})
		return
	}

	log.Printf("File served successfully: %s", filePath)
}
