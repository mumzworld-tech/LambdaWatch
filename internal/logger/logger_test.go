package logger

import (
	"os"
	"testing"

	"github.com/Sami-AlEsh/lambdawatch/internal/buffer"
)

func TestInit_AppNameFromAPP_NAME(t *testing.T) {
	os.Setenv("APP_NAME", "my-app")
	os.Setenv("SERVICE_NAME", "my-service")
	defer os.Unsetenv("APP_NAME")
	defer os.Unsetenv("SERVICE_NAME")

	Init()
	if appName != "my-app" {
		t.Errorf("expected my-app, got %s", appName)
	}
}

func TestInit_AppNameFallsBackToServiceName(t *testing.T) {
	os.Unsetenv("APP_NAME")
	os.Setenv("SERVICE_NAME", "my-service")
	defer os.Unsetenv("SERVICE_NAME")

	Init()
	if appName != "my-service" {
		t.Errorf("expected my-service, got %s", appName)
	}
}

func TestInit_Environment(t *testing.T) {
	os.Setenv("NODE_ENV", "production")
	defer os.Unsetenv("NODE_ENV")

	Init()
	if environment != "production" {
		t.Errorf("expected production, got %s", environment)
	}
}

func TestInit_EnvironmentDefault(t *testing.T) {
	os.Unsetenv("NODE_ENV")

	Init()
	if environment != "unknown" {
		t.Errorf("expected unknown, got %s", environment)
	}
}

func TestSetBuffer_WritesToBuffer(t *testing.T) {
	buf := buffer.New(100)
	SetBuffer(buf)
	defer SetBuffer(nil)

	Info("test message")

	if buf.Len() != 1 {
		t.Errorf("expected 1 entry in buffer, got %d", buf.Len())
	}
	entries := buf.Flush(1)
	if entries[0].Type != "extension" {
		t.Errorf("expected type extension, got %s", entries[0].Type)
	}
}

func TestSetBuffer_NilBufferNoWrite(t *testing.T) {
	SetBuffer(nil)
	// Should not panic
	Info("test message")
}

func TestLogLevels(t *testing.T) {
	buf := buffer.New(100)
	SetBuffer(buf)
	defer SetBuffer(nil)

	Info("info msg")
	Debug("debug msg")
	Error("error msg")

	if buf.Len() != 3 {
		t.Errorf("expected 3 entries, got %d", buf.Len())
	}
}

func TestLogFormatFunctions(t *testing.T) {
	buf := buffer.New(100)
	SetBuffer(buf)
	defer SetBuffer(nil)

	Infof("hello %s", "world")
	Debugf("count %d", 42)
	Errorf("err: %v", "fail")

	if buf.Len() != 3 {
		t.Errorf("expected 3 entries, got %d", buf.Len())
	}
}
