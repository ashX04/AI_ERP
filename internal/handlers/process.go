package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ashX04/new_website/internal/utils"
	"github.com/xuri/excelize/v2"
)

// ProcessImage handles sending the image to Azure Vision API and processing the response with OpenAI
func ProcessImage(filePath string, userID string) (string, error) {
	// Send the image to the Azure Vision API
	resp, err := utils.SendImageToAPI(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to send image to API: %w", err)
	}
	time.Sleep(2 * time.Second)

	// Handle the API response
	responseData, err := utils.HandleAPIResponse(resp)
	if err != nil {
		return "", fmt.Errorf("failed to handle API response: %w", err)
	}

	// Parse the JSON response
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(responseData), &result); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		// If parsing fails, try sending the raw response to OpenAI
		//csvData, err := utils.SendJSONToOpenAI(responseData)
		// if err != nil {
		// 	return "", fmt.Errorf("failed to process text with OpenAI: %w", err)
		// }
		csvData := "error true"
		return csvData, nil
	}

	// Extract text from the Azure Vision API response
	var extractedText string
	if analyzeResult, ok := result["analyzeResult"].(map[string]interface{}); ok {
		if readResults, ok := analyzeResult["readResults"].([]interface{}); ok {
			for _, page := range readResults {
				if pageObj, ok := page.(map[string]interface{}); ok {
					if lines, ok := pageObj["lines"].([]interface{}); ok {
						for _, line := range lines {
							if lineObj, ok := line.(map[string]interface{}); ok {
								if text, ok := lineObj["text"].(string); ok {
									extractedText += text + " "
								}
							}
						}
					}
				}
			}
		}
	}

	// If no text was extracted, use the raw response
	if extractedText == "" {
		extractedText = responseData
	}

	log.Printf("Extracted Text: %s", extractedText)

	// Process the extracted text with OpenAI
	csvData, err := utils.SendJSONToOpenAI(extractedText)
	if err != nil {
		return "", fmt.Errorf("failed to process text with OpenAI: %w", err)
	}
	log.Printf("CSV Data: %s", csvData)

	// Extract text between <*> tags
	startIndex := strings.Index(csvData, "<*>")
	endIndex := strings.LastIndex(csvData, "<*>")

	if startIndex != -1 && endIndex != -1 && endIndex > startIndex {
		csvData = csvData[startIndex+3 : endIndex]
	} else {
		log.Printf("Warning: <*> tags not found in csvData")
	}
	// Remove all <*> from csvData
	csvData = strings.ReplaceAll(csvData, "<*>", "")
	// Create a new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing Excel file: %v", err)
		}
	}()

	// Set the active sheet
	sheetName := "Sheet1"
	index, err := f.GetSheetIndex(sheetName)
	if err != nil {
		log.Printf("Error getting sheet index: %v", err)
		return "", fmt.Errorf("failed to get sheet index: %w", err)
	}
	f.SetActiveSheet(index)

	// Split the CSV data into rows
	rows := strings.Split(strings.TrimSpace(csvData), "\n")

	// Write each row to the Excel file
	for i, row := range rows {
		cols := strings.Split(row, ",")
		for j, col := range cols {
			cell, err := excelize.CoordinatesToCellName(j+1, i+1)
			if err != nil {
				log.Printf("Error converting coordinates to cell name: %v", err)
				continue
			}
			f.SetCellValue(sheetName, cell, strings.TrimSpace(col))
		}
	}

	// Save the Excel file
	fileName := fmt.Sprintf("output_%d.xlsx", time.Now().Unix())
	if err := f.SaveAs(fileName); err != nil {
		log.Printf("Error saving Excel file: %v", err)
		return "", fmt.Errorf("failed to save Excel file: %w", err)
	}
	// Prepare file data for PocketBase
	fileData := &bytes.Buffer{}
	writer := multipart.NewWriter(fileData)

	// Add user ID field
	if err := writer.WriteField("user", userID); err != nil {
		log.Printf("Error writing user field: %v", err)
		return "", fmt.Errorf("failed to write user field: %w", err)
	}

	// Add Excel file
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("Error opening Excel file: %v", err)
		return "", fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("excel", fileName)
	if err != nil {
		log.Printf("Error creating form file: %v", err)
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		log.Printf("Error copying file contents: %v", err)
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	writer.Close()

	// Send request to PocketBase
	req, err := http.NewRequest("POST", "http://localhost:8090/api/collections/excel_files/records", fileData)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	pbResp, err := client.Do(req) // Changed variable name to pbResp
	if err != nil {
		log.Printf("Error uploading to PocketBase: %v", err)
		return "", fmt.Errorf("failed to upload to PocketBase: %w", err)
	}
	defer pbResp.Body.Close()

	if pbResp.StatusCode != http.StatusOK {
		log.Printf("PocketBase upload failed with status: %d", pbResp.StatusCode)
		return "", fmt.Errorf("PocketBase upload failed with status: %d", pbResp.StatusCode)
	}

	log.Printf("Excel file saved as: %s", fileName)
	csvData = fileName // Update csvData to return the file name

	return csvData, nil
}
