package utils

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

// SendImageToAPI sends an image to the Azure Vision API
func SendImageToAPI(imagePath string) (*http.Response, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	// Get the API token from the environment variable
	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		return nil, fmt.Errorf("API token not found in environment")
	}

	// Open the image file
	fmt.Println("Image Path:", imagePath)
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	fmt.Println("Absolute Image Path:", absPath)

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := bufio.NewReader(file)
	// Create the API request
	apiURL := "https://centralindia.api.cognitive.microsoft.com/vision/v3.2/read/analyze?model-version=latest" // Replace with your actual API URL
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add necessary headers
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("Ocp-Apim-Subscription-Key", apiToken)

	// Send the request using http.Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

// HandleAPIResponse processes the response from the Azure Vision API
func HandleAPIResponse(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	err := godotenv.Load()
	if err != nil {
		return "error", fmt.Errorf("failed to load .env file: %w", err)
	}

	// Get the API token from the environment variable
	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		return "error", fmt.Errorf("API token not found in environment")
	}

	fmt.Println(resp.Header)
	contentType := resp.Header.Get("Content-Type")
	fmt.Println("Content-Type:", contentType)

	apiURL := resp.Header.Get("Operation-Location")
	fmt.Println("Operation-Location:", apiURL)

	time.Sleep(2 * time.Second)

	response, err := MakeGetRequestWithAuth(apiURL, apiToken)
	if err != nil {
		fmt.Println("Error:", err)
		return "error", err
	}
	return response, nil
}

// MakeGetRequestWithAuth makes a GET request with authentication
func MakeGetRequestWithAuth(apiURL string, token string) (string, error) {
	// Create a new HTTP GET request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", token) // Bearer token

	// Create an HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-200 status codes (optional)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-OK status code: %d, body: %s", resp.StatusCode, bodyBytes)
	}

	// Return the response body as a string
	return string(bodyBytes), nil
}

// SendJSONToOpenAI sends JSON data to OpenAI and returns the processed data
func SendJSONToOpenAI(data string) (string, error) {
	prompt := fmt.Sprintf("Use this to make a table %s now convert it to a CSV in this column order: Serial.no.,Quantity.,Pack,HSN no.,Product_name,batch no.,Expiry date,MRP,S.Rate(selling rate),GST,CGST,SGST,Amount (GST is cgst = sgst). Ignore other data and only give the CSV and nothing else. Also put <*> at the start and end of the CSV.", data)

	err := godotenv.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load .env file: %w", err)
	}

	c := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	ctx := context.Background()
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT4oMini,
		MaxTokens: 1500,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
