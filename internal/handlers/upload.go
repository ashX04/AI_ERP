package handlers

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
)

// UploadImage handles the uploading of multiple images
func UploadImage(c *gin.Context) {
	// Parse the uploaded file
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to parse form: %v", err)
		return
	}

	files := form.File["files"]
	if len(files) > 10 {
		c.String(http.StatusBadRequest, "You can upload up to 10 files at a time.")
		return
	}

	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()

			filename := filepath.Base(file.Filename)
			filePath := fmt.Sprintf("./uploads/%s", filename)
			if err := c.SaveUploadedFile(file, filePath); err != nil {
				log.Printf("Failed to upload file: %v", err)
				return
			}

			log.Printf("File uploaded successfully: %s", filename)

			// Process the image
			csvData, err := ProcessImage(filePath)
			if err != nil {
				log.Printf("Failed to process image: %v", err)
				return
			}

			log.Printf("CSV Data: %s", csvData)

		}(file)
	}

	wg.Wait()
	c.String(http.StatusOK, "Files uploaded and processed successfully.")
}
