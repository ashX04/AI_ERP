package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ashX04/new_website/internal/utils"
	"github.com/xuri/excelize/v2"
)

// ProcessImage handles sending the image to Azure Vision API and processing the response with OpenAI
func ProcessImage(filePath string) (string, error) {
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

	log.Printf("Excel file saved as: %s", fileName)
	csvData = fileName // Update csvData to return the file name

	return csvData, nil
}
