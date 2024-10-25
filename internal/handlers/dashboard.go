package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	sessions "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ExcelFile struct {
	ID          string `json:"id"`
	User        string `json:"user"`
	ExcelPath   string `json:"excel"`
	SourceImage string `json:"source_image"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
	FileName    string `json:"-"`
	ImagePath   string `json:"-"`
	CreatedAt   string `json:"-"` // Add this field for parsed time
}

type PocketBaseResponse struct {
	Page       int         `json:"page"`
	PerPage    int         `json:"perPage"`
	TotalItems int         `json:"totalItems"`
	Items      []ExcelFile `json:"items"`
}

func ShowDashboard(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Add debug logging for userID
	log.Printf("Current userID: %v", userID)

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

	// Filter files for current user
	var userFiles []ExcelFile
	for _, file := range pbResp.Items {
		if file.User == userID.(string) {
			userFiles = append(userFiles, file)
		}
	}

	// Log the number of user files
	log.Printf("Number of files for user %v: %d", userID, len(userFiles))

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

func PreviewImage(c *gin.Context) {
	fileID := c.Param("id")
	log.Printf("Attempting to preview image for file with ID: %s", fileID)

	// Fetch excel file details from PocketBase
	excelResp, err := http.Get("http://localhost:8090/api/collections/excel_files/records/" + fileID)
	if err != nil {
		log.Printf("Error fetching excel file details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch excel file details"})
		return
	}
	defer excelResp.Body.Close()

	var excelFile ExcelFile
	if err := json.NewDecoder(excelResp.Body).Decode(&excelFile); err != nil {
		log.Printf("Error parsing excel file details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse excel file details"})
		return
	}

	// Fetch image details from PocketBase
	imageResp, err := http.Get("http://localhost:8090/api/collections/images/records/" + excelFile.SourceImage)
	if err != nil {
		log.Printf("Error fetching image details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch image details"})
		return
	}
	defer imageResp.Body.Close()

	var imageFile struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(imageResp.Body).Decode(&imageFile); err != nil {
		log.Printf("Error parsing image details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse image details"})
		return
	}

	// Construct the image file path
	pbImageStoragePath := "D:\\codes\\golang\\basedatabase\\pb_data\\storage\\ja5zwzf3eiqb4tc"
	imagePath := filepath.Join(pbImageStoragePath, imageFile.ID, imageFile.Name)
	log.Printf("Constructed image path: %s", imagePath)

	// Check if image exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log.Printf("Image not found: %s", imagePath)
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Read the image file
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		log.Printf("Error reading image file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image file"})
		return
	}

	contentType := getContentType(imagePath)
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, imageData)

	log.Printf("Image path: %s, Content-Type: %s", imagePath, contentType)
}

func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}
