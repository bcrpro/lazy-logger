package lazylogger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Logger is the core struct for LazyLogger
// storage: 0: local, 1: loki, 2: both
type Logger struct {
	lokiURL string
	jobName string
	storage int
}

var (
	instance *Logger
	once     sync.Once
)

// Init initializes the global logger instance
func Init(lokiURL, jobName string,  storage int) {
	once.Do(func() {
		instance = &Logger{
			lokiURL: lokiURL,
			jobName: jobName,
			storage: storage,
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
	if l.storage == 0 {
		log.Printf("[%s] %s", level, message)
		return
	}
	if l.storage == 2 {
		log.Printf("[%s] %s", level, message)
	}
	payload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": map[string]string{
					"job":   l.jobName,
					"level": level,
				},
				"values": [][]string{
					{
						fmt.Sprintf("%d", time.Now().UnixNano()),
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