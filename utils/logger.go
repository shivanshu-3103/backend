package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	successLogger *log.Logger
	errorLogger   *log.Logger
)

func init() {
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		fmt.Printf("Failed to create logs directory: %v\n", err)
		return
	}

	successFile, err := os.OpenFile(filepath.Join("logs", "access.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open access.log: %v\n", err)
	} else {
		successLogger = log.New(successFile, "SUCCESS: ", log.Ldate|log.Ltime)
	}

	errorFile, err := os.OpenFile(filepath.Join("logs", "error.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open error.log: %v\n", err)
	} else {
		errorLogger = log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

// LogSuccess logs a successful event or info message to access.log
func LogSuccess(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if successLogger != nil {
		successLogger.Println(msg)
	} else {
		fmt.Printf("[%s] SUCCESS: %s\n", time.Now().Format(time.RFC3339), msg)
	}
}

// LogInfo is an alias for LogSuccess
func LogInfo(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if successLogger != nil {
		successLogger.Println(msg)
	} else {
		fmt.Printf("[%s] INFO: %s\n", time.Now().Format(time.RFC3339), msg)
	}
}

// LogError logs an error to error.log
func LogError(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if errorLogger != nil {
		errorLogger.Println(msg)
	} else {
		fmt.Printf("[%s] ERROR: %s\n", time.Now().Format(time.RFC3339), msg)
	}
}

// LogFatal logs an error to error.log and then calls os.Exit(1)
func LogFatal(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if errorLogger != nil {
		errorLogger.Println(msg)
	} else {
		fmt.Printf("[%s] FATAL: %s\n", time.Now().Format(time.RFC3339), msg)
	}
	os.Exit(1)
}
