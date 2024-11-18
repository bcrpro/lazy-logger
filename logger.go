package lazylogger

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// Logger is the core struct for LazyLogger
type Logger struct {
	lokiURL string
	jobName string
}

var (
	instance *Logger
	once     sync.Once
)

// Init initializes the global logger instance
func Init(lokiURL, jobName string) {
	once.Do(func() {
		instance = &Logger{
			lokiURL: lokiURL,
			jobName: jobName,
		}
	})
}

// GetLogger returns the singleton instance of Logger
func GetLogger() *Logger {
	if instance == nil {
		log.Fatal("LazyLogger not initialized. Call lazylogger.Init() first.")
	}
	return instance
}

// pushToLoki sends a log entry to the Loki server
func (l *Logger) pushToLoki(level, message string) {
	payload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": map[string]string{
					"job":   l.jobName,
					"level": level,
				},
				"values": [][]string{
					{
						time.Now().Format("20060102150405") + "000000000",
						message,
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal log payload: %v", err)
		return
	}

	resp, err := http.Post(l.lokiURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Failed to send log to Loki: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		log.Printf("Unexpected Loki response status: %d", resp.StatusCode)
	}
}

// Info logs an informational message
func (l *Logger) Info(message string) {
	l.pushToLoki("info", message)
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	l.pushToLoki("warn", message)
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.pushToLoki("error", message)
}