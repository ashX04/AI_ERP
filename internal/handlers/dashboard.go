package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"time"

	"github.com/ashX04/new_website/internal/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type DashboardData struct {
	Title string
	Files []FileData
	Error string
}

type FileData struct {
	ID        string
	Created   string
	Image     string
	ExcelFile string
}

type PocketBaseResponse struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	Items      []struct {
		ID      string `json:"id"`
		Created string `json:"created"`
		Excel   string `json:"excel"`
		Image   string `json:"image"`
		User    string `json:"user"`
	} `json:"items"`
}

func ShowDashboard(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("userID")

	if userID == nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"error": "Please login first",
		})
		return
	}

	// Sanitize user ID
	userIDStr := html.EscapeString(fmt.Sprintf("%v", userID))

	// Create safe URL with proper escaping
	baseURL := "http://127.0.0.1:8090/api/collections/excel_files/records"
	params := url.Values{}
	params.Add("filter", fmt.Sprintf("(user='%s')", url.QueryEscape(userIDStr)))

	// Make HTTP request with secure client and timeout
	resp, err := utils.SecureClient.Get(baseURL + "?" + params.Encode())
	if err != nil {
		c.HTML(http.StatusOK, "dashboard.html", DashboardData{
			Title: "Dashboard",
			Error: "Failed to fetch files",
			Files: []FileData{},
		})
		return
	}
	defer resp.Body.Close()

	// Parse response
	var pbResp PocketBaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&pbResp); err != nil {
		c.HTML(http.StatusOK, "dashboard.html", DashboardData{
			Title: "Dashboard",
			Error: "Failed to parse response: " + err.Error(),
			Files: []FileData{},
		})
		return
	}

	var files []FileData
	for _, item := range pbResp.Items {
		// Double check that the file belongs to the user
		if item.User != userID.(string) {
			continue
		}

		createdTime, err := time.Parse("2006-01-02 15:04:05.999Z", item.Created)
		if err != nil {
			continue // Skip this item if time parsing fails
		}

		fileData := FileData{
			ID:      item.ID,
			Created: createdTime.Format("2006-01-02 15:04:05.000"),
		}

		// Set Excel file URL
		if item.Excel != "" {
			fileData.ExcelFile = fmt.Sprintf("http://127.0.0.1:8090/api/files/excel_files/%s/%s",
				item.ID,
				item.Excel)
		}

		// Set Image URL
		if item.Image != "" {
			fileData.Image = fmt.Sprintf("http://127.0.0.1:8090/api/files/excel_files/%s/%s",
				item.ID,
				item.Image)
		}

		files = append(files, fileData)
	}

	c.HTML(http.StatusOK, "dashboard.html", DashboardData{
		Title: "Dashboard",
		Files: files,
	})
}

// DownloadFile handles file downloads
func DownloadFile(c *gin.Context) {
	id := c.Param("id")
	fileURL := fmt.Sprintf("http://127.0.0.1:8090/api/files/excel_files/%s", id)
	c.Redirect(http.StatusFound, fileURL)
}

// DeleteFile handles file deletion
func DeleteFile(c *gin.Context) {
	id := c.Param("id")

	// Validate file ID
	if !utils.ValidateFileID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("userID")

	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Create safe URL using the sanitized ID directly
	safeURL, err := utils.SanitizeURL(fmt.Sprintf("http://127.0.0.1:8090/api/collections/excel_files/records/%s", url.QueryEscape(id)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	// Use secure client for requests
	resp, err := utils.SecureClient.Get(safeURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify file ownership"})
		return
	}
	defer resp.Body.Close()

	var fileRecord struct {
		User string `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&fileRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify file ownership"})
		return
	}

	// Check if the file belongs to the user
	if fileRecord.User != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to delete this file"})
		return
	}

	// Use secure client for deletion
	req, err := http.NewRequest("DELETE", safeURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create delete request"})
		return
	}

	deleteResp, err := utils.SecureClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}
	defer deleteResp.Body.Close()

	c.Status(http.StatusOK)
}

// PreviewImage handles image preview
func PreviewImage(c *gin.Context) {
	id := c.Param("id")

	// Validate file ID
	if !utils.ValidateFileID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("userID")

	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Create safe URL
	url, err := utils.SanitizeURL(fmt.Sprintf("http://127.0.0.1:8090/api/collections/excel_files/records/%s", url.QueryEscape(id)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	// Use secure client
	resp, err := utils.SecureClient.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify file ownership"})
		return
	}
	defer resp.Body.Close()

	var fileRecord struct {
		User string `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&fileRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify file ownership"})
		return
	}

	// Check if the file belongs to the user
	if fileRecord.User != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to view this file"})
		return
	}

	imageURL := fmt.Sprintf("http://127.0.0.1:8090/api/files/excel_files/%s", id)
	c.Redirect(http.StatusFound, imageURL)
}
