package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Sami-AlEsh/lambdawatch/internal/buffer"
)

var (
	appName     string
	environment string
	logBuffer   *buffer.Buffer
)

func Init() {
	appName = os.Getenv("APP_NAME")
	if appName == "" {
		appName = os.Getenv("SERVICE_NAME")
	}
	environment = os.Getenv("NODE_ENV")
	if environment == "" {
		environment = "unknown"
	}
}

// SetBuffer sets the buffer for extension logs to be written directly
// This is necessary because Telemetry API doesn't capture logs from the same extension
func SetBuffer(buf *buffer.Buffer) {
	logBuffer = buf
}

type logEntry struct {
	Level       string `json:"level"`
	Timestamp   string `json:"timestamp"`
	AppName     string `json:"app_name"`
	Environment string `json:"environment"`
	Context     string `json:"context"`
	Message     string `json:"message"`
}

func log(level, msg string) {
	entry := logEntry{
		Level:       level,
		Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
		AppName:     appName,
		Environment: environment,
		Context:     "LambdaWatch",
		Message:     msg,
	}
	b, _ := json.Marshal(entry)
	logLine := string(b)

	// Always write to stdout for CloudWatch
	fmt.Println(logLine)

	// Also write directly to buffer for Loki (Telemetry API won't capture our own logs)
	if logBuffer != nil {
		logBuffer.Add(buffer.LogEntry{
			Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			Message:   logLine,
			Type:      "extension",
		})
		// Signal that logs are ready for flushing
		logBuffer.SignalReady()
	}
}

func Info(msg string)                   { log("info", msg) }
func Debug(msg string)                  { log("debug", msg) }
func Error(msg string)                  { log("error", msg) }
func Infof(format string, a ...any)     { log("info", fmt.Sprintf(format, a...)) }
func Debugf(format string, a ...any)    { log("debug", fmt.Sprintf(format, a...)) }
func Errorf(format string, a ...any)    { log("error", fmt.Sprintf(format, a...)) }
func Fatalf(format string, a ...any)    { log("fatal", fmt.Sprintf(format, a...)); os.Exit(1) }
func Fatal(msg string)                  { log("fatal", msg); os.Exit(1) }
