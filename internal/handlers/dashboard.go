package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	Image     string // Changed from SourceImage
	ExcelFile string
}

type PocketBaseResponse struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	Items      []struct {
		ID      string `json:"id"`
		Created string `json:"created"`
		Excel   string `json:"excel"` // Updated field name
		Image   string `json:"image"` // New field name
		User    string `json:"user"`
	} `json:"items"`
}

func ShowDashboard(c *gin.Context) {
	// Make HTTP request to PocketBase API - removed expand parameter as it's no longer needed
	resp, err := http.Get("http://127.0.0.1:8090/api/collections/excel_files/records")
	if err != nil {
		c.HTML(http.StatusOK, "dashboard.html", DashboardData{
			Title: "Dashboard",
			Error: "Failed to fetch files: " + err.Error(),
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
		createdTime, err := time.Parse("2006-01-02 15:04:05.999Z", item.Created)
		if err != nil {
			c.HTML(http.StatusOK, "dashboard.html", DashboardData{
				Title: "Dashboard",
				Error: "Failed to parse time: " + err.Error(),
				Files: []FileData{},
			})
			return
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

	// Make DELETE request to PocketBase
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("http://127.0.0.1:8090/api/collections/excel_files/records/%s", id), nil)
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	c.Status(http.StatusOK)
}

// PreviewImage handles image preview
func PreviewImage(c *gin.Context) {
	id := c.Param("id")
	// Updated to use excel_files collection instead of images
	imageURL := fmt.Sprintf("http://127.0.0.1:8090/api/files/excel_files/%s", id)
	c.Redirect(http.StatusFound, imageURL)
}
