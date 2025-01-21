package logger

import (
	"log"
	"os"
)

// Logger instance
var logger *log.Logger

// Init initializes the logger
func Init(logFilePath string) error {
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		return err
	}

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

func Info(format string, v ...interface{}) {
	logger.Printf("[INFO] "+format, v...)
}

func Error(format string, v ...interface{}) {
	logger.Printf("[ERROR] "+format, v...)
}

func Fatalf(format string, v ...interface{}) {
	logger.Printf("[FATAL] "+format, v...)
	os.Exit(1)
}
