package logger

import (
	"fmt"
	"log"
	"os"
)

var (
	logFile *os.File
	logger  *log.Logger
)

func init() {
	var err error
	logFile, err = os.OpenFile("resh.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
	}
	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
}

func Debug(msg string, args ...interface{}) {
	logger.Printf("[DEBUG] "+msg, args...)
}

func Info(msg string, args ...interface{}) {
	logger.Printf("[INFO] "+msg, args...)
}

func Error(msg string, args ...interface{}) {
	logger.Printf("[ERROR] "+msg, args...)
}

func JSON(label string, data interface{}) {
	logger.Printf("[JSON] %s: %#v", label, data)
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}
