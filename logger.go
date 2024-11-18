package logger

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

type Logger struct {
	lokiURL string
	jobName string
}

var (
	instance *Logger
	once     sync.Once
)

// Initialize the global logger instance
func Init(lokiURL, jobName string) {
	once.Do(func() {
		instance = &Logger{
			lokiURL: lokiURL,
			jobName: jobName,
		}
	})
}

// Get the global logger instance
func GetLogger() *Logger {
	if instance == nil {
		log.Fatal("Logger not initialized. Call logger.Init() first.")
	}
	return instance
}

// pushToLoki sends a log message to Loki
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
						// Timestamp in nanoseconds
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

// Public log methods
func (l *Logger) Info(message string) {
	l.pushToLoki("info", message)
}

func (l *Logger) Warn(message string) {
	l.pushToLoki("warn", message)
}

func (l *Logger) Error(message string) {
	l.pushToLoki("error", message)
}