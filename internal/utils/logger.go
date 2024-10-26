package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

var Logger *log.Logger

func init() {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	// Open log file with current date
	logFile, err := os.OpenFile(
		fmt.Sprintf("logs/app_%s.log", time.Now().Format("2006-01-02")),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	Logger = log.New(logFile, "", log.Ldate|log.Ltime)
}
