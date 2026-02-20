package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	file       *os.File
	infoLog    *log.Logger
	errorLog   *log.Logger
	successLog *log.Logger
}

func NewLogger(filename string) (*Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:       file,
		infoLog:    log.New(file, "[INFO] ", log.Ldate|log.Ltime),
		errorLog:   log.New(file, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
		successLog: log.New(file, "[SUCCESS] ", log.Ldate|log.Ltime),
	}, nil
}

func (l *Logger) Close() error {
	return l.file.Close()
}

func (l *Logger) Info(message string) {
	l.infoLog.Println(message)
	fmt.Println("[INFO]", message)
}

func (l *Logger) Error(message string, err error) {
	l.errorLog.Printf("%s: %v\n", message, err)
	fmt.Printf("[ERROR] %s: %v\n", message, err)
}

func (l *Logger) Success(message string) {
	l.successLog.Println(message)
	fmt.Println("[SUCCESS]", message)
}

func (l *Logger) LogScrapingSession(totalListings int, duration time.Duration) {
	msg := fmt.Sprintf("Scraped %d listings in %v", totalListings, duration)
	l.successLog.Println(msg)
	fmt.Println("[SESSION]", msg)
}