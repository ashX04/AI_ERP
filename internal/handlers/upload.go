package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
)

// UploadImage handles the uploading of multiple images
func UploadImage(c *gin.Context) {
	// Get user ID from session first
	session, err := store.Get(c.Request, "session-name")
	if err != nil {
		c.HTML(http.StatusUnauthorized, "upload.html", gin.H{
			"error": "Session error, please login again",
		})
		return
	}

	userID, ok := session.Values["userID"].(string)
	if !ok {
		c.HTML(http.StatusUnauthorized, "upload.html", gin.H{
			"error": "User not authenticated, please login again",
		})
		return
	}

	// Parse the uploaded file
	form, err := c.MultipartForm()
	if err != nil {
		c.HTML(http.StatusBadRequest, "upload.html", gin.H{
			"error": fmt.Sprintf("Failed to parse form: %v", err),
		})
		return
	}

	files := form.File["files"]
	if len(files) > 10 {
		c.HTML(http.StatusBadRequest, "upload.html", gin.H{
			"error": "You can upload up to 10 files at a time.",
		})
		return
	}

	var wg sync.WaitGroup
	errorChan := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()

			filename := filepath.Base(file.Filename)
			filePath := fmt.Sprintf("./uploads/%s", filename)
			if err := c.SaveUploadedFile(file, filePath); err != nil {
				errorChan <- fmt.Errorf("failed to save file %s: %v", filename, err)
				return
			}

			// Prepare file data for PocketBase
			fileData := &bytes.Buffer{}
			writer := multipart.NewWriter(fileData)

			// Add user ID field
			if err := writer.WriteField("user", userID); err != nil {
				errorChan <- fmt.Errorf("failed to write user field for %s: %v", filename, err)
				return
			}

			// Add image file
			part, err := writer.CreateFormFile("image", filename)
			if err != nil {
				errorChan <- fmt.Errorf("failed to create form file for %s: %v", filename, err)
				return
			}

			// Open and copy file contents
			src, err := file.Open()
			if err != nil {
				errorChan <- fmt.Errorf("failed to open file %s: %v", filename, err)
				return
			}
			defer src.Close()

			if _, err := io.Copy(part, src); err != nil {
				errorChan <- fmt.Errorf("failed to copy file contents for %s: %v", filename, err)
				return
			}

			writer.Close()

			// Send request to PocketBase
			req, err := http.NewRequest("POST", "http://localhost:8090/api/collections/images/records", fileData)
			if err != nil {
				errorChan <- fmt.Errorf("failed to create request for %s: %v", filename, err)
				return
			}

			req.Header.Set("Content-Type", writer.FormDataContentType())

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				errorChan <- fmt.Errorf("failed to upload %s to PocketBase: %v", filename, err)
				return
			}
			defer resp.Body.Close()

			// Read and parse the response to get imageID
			var pbResponse struct {
				Id string `json:"id"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&pbResponse); err != nil {
				errorChan <- fmt.Errorf("failed to decode PocketBase response for %s: %v", filename, err)
				return
			}

			// Process the image with the obtained imageID
			csvData, err := ProcessImage(filePath, userID, pbResponse.Id)
			if err != nil {
				errorChan <- fmt.Errorf("failed to process image %s: %v", filename, err)
				return
			}

			log.Printf("File %s processed successfully. CSV Data: %s", filename, csvData)

		}(file)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errorChan)

	// Collect any errors
	var errors []string
	for err := range errorChan {
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		c.HTML(http.StatusInternalServerError, "upload.html", gin.H{
			"error": fmt.Sprintf("Some files failed to upload: %v", errors),
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/dashboard")
}
