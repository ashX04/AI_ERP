package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ashX04/new_website/internal/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type DashboardData struct {
	Title      string
	FileGroups []FileGroup
	Error      string
}

type FileGroup struct {
	Date  string
	Files []FileData
}

type FileData struct {
	ID        string
	Created   string
	CreatedAt time.Time
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
			Title:      "Dashboard",
			Error:      "Failed to fetch files",
			FileGroups: []FileGroup{},
		})
		return
	}
	defer resp.Body.Close()

	// Parse response
	var pbResp PocketBaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&pbResp); err != nil {
		c.HTML(http.StatusOK, "dashboard.html", DashboardData{
			Title:      "Dashboard",
			Error:      "Failed to parse response: " + err.Error(),
			FileGroups: []FileGroup{},
		})
		return
	}

	// Create files slice
	var files []FileData
	for _, item := range pbResp.Items {
		// Double check that the file belongs to the user
		if item.User != userID.(string) {
			continue
		}

		createdTime, err := time.Parse("2006-01-02 15:04:05.999Z", item.Created)
		if err != nil {
			continue
		}

		fileData := FileData{
			ID:        item.ID,
			Created:   createdTime.Format("2006-01-02 15:04:05"),
			CreatedAt: createdTime, // Store the time.Time for sorting
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

	// Sort files by date (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].CreatedAt.After(files[j].CreatedAt)
	})

	// Group files by date
	fileGroups := groupFilesByDate(files)

	c.HTML(http.StatusOK, "dashboard.html", DashboardData{
		Title:      "Dashboard",
		FileGroups: fileGroups,
	})
}

// Helper function to group files by date
func groupFilesByDate(files []FileData) []FileGroup {
	groups := make(map[string][]FileData)

	for _, file := range files {
		date := file.CreatedAt.Format("January 2, 2006") // Format date as "Month Day, Year"
		groups[date] = append(groups[date], file)
	}

	// Convert map to sorted slice
	var fileGroups []FileGroup
	for date, files := range groups {
		fileGroups = append(fileGroups, FileGroup{
			Date:  date,
			Files: files,
		})
	}

	// Sort groups by date (newest first)
	sort.Slice(fileGroups, func(i, j int) bool {
		dateI, _ := time.Parse("January 2, 2006", fileGroups[i].Date)
		dateJ, _ := time.Parse("January 2, 2006", fileGroups[j].Date)
		return dateI.After(dateJ)
	})

	return fileGroups
}

// DownloadFile handles file downloads
func DownloadFile(c *gin.Context) {
	id := c.Param("id")

	// Get the file info first
	infoURL := fmt.Sprintf("http://127.0.0.1:8090/api/collections/excel_files/records/%s", id)
	resp, err := utils.SecureClient.Get(infoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info"})
		return
	}
	defer resp.Body.Close()

	var fileInfo struct {
		Excel string `json:"excel"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse file info"})
		return
	}

	// Construct the correct download URL with the filename
	fileURL := fmt.Sprintf("http://127.0.0.1:8090/api/files/excel_files/%s/%s", id, fileInfo.Excel)
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

	// Get the file info first
	infoURL := fmt.Sprintf("http://127.0.0.1:8090/api/collections/excel_files/records/%s", id)
	resp, err := utils.SecureClient.Get(infoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info"})
		return
	}
	defer resp.Body.Close()

	var fileInfo struct {
		User  string `json:"user"`
		Image string `json:"image"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse file info"})
		return
	}

	// Check if the file belongs to the user
	if fileInfo.User != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to view this file"})
		return
	}

	// Construct the correct image URL with the filename
	imageURL := fmt.Sprintf("http://127.0.0.1:8090/api/files/excel_files/%s/%s", id, fileInfo.Image)
	c.Redirect(http.StatusFound, imageURL)
}

// Add this new function to handle multiple downloads
func DownloadMultipleFiles(c *gin.Context) {
	// Get file IDs from query parameter
	fileIDs := c.Query("files")
	if fileIDs == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files selected"})
		return
	}

	// Get user session
	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Split file IDs
	ids := strings.Split(fileIDs, ",")

	// Create a zip file
	tmpfile, err := os.CreateTemp("", "download-*.zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary file"})
		return
	}
	defer os.Remove(tmpfile.Name())

	zipWriter := zip.NewWriter(tmpfile)
	defer zipWriter.Close()

	// Process each file
	for _, id := range ids {
		// Validate file ID
		if !utils.ValidateFileID(id) {
			continue
		}

		// Get file info
		infoURL := fmt.Sprintf("http://127.0.0.1:8090/api/collections/excel_files/records/%s", id)
		resp, err := utils.SecureClient.Get(infoURL)
		if err != nil {
			continue
		}

		var fileInfo struct {
			User  string `json:"user"`
			Excel string `json:"excel"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Verify ownership
		if fileInfo.User != userID.(string) {
			continue
		}

		// Download the file
		fileURL := fmt.Sprintf("http://127.0.0.1:8090/api/files/excel_files/%s/%s", id, fileInfo.Excel)
		fileResp, err := utils.SecureClient.Get(fileURL)
		if err != nil {
			continue
		}

		// Create file in zip
		f, err := zipWriter.Create(fileInfo.Excel)
		if err != nil {
			fileResp.Body.Close()
			continue
		}

		// Copy file content to zip
		_, err = io.Copy(f, fileResp.Body)
		fileResp.Body.Close()
		if err != nil {
			continue
		}
	}

	// Close the zip writer before reading
	zipWriter.Close()

	// Set headers for download
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=excel_files.zip")

	// Send the file
	c.File(tmpfile.Name())
}
