package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var (
	appName     string
	environment string
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
	fmt.Println(string(b))
}

func Info(msg string)                   { log("info", msg) }
func Debug(msg string)                  { log("debug", msg) }
func Error(msg string)                  { log("error", msg) }
func Infof(format string, a ...any)     { log("info", fmt.Sprintf(format, a...)) }
func Debugf(format string, a ...any)    { log("debug", fmt.Sprintf(format, a...)) }
func Errorf(format string, a ...any)    { log("error", fmt.Sprintf(format, a...)) }
func Fatalf(format string, a ...any)    { log("fatal", fmt.Sprintf(format, a...)); os.Exit(1) }
func Fatal(msg string)                  { log("fatal", msg); os.Exit(1) }
