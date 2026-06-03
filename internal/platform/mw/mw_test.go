package mw

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoggerStructuredWritesRequestMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	restoreLogger(&logs, t)

	r := gin.New()
	r.Use(RequestID(), func(c *gin.Context) {
		c.Set(ContextUserIDKey, int64(42))
		c.Next()
	}, LoggerStructured())
	r.GET("/stations", func(c *gin.Context) {
		c.String(http.StatusCreated, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/stations?limit=10", nil)
	req.Header.Set(HeaderRequestID, "req-1")
	req.Header.Set("User-Agent", "test-agent")
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	entry := firstLogEntry(t, logs.String())
	assertEqual(t, entry["msg"], "http_request")
	assertEqual(t, entry["level"], "INFO")
	assertEqual(t, entry["request_id"], "req-1")
	assertEqual(t, entry["method"], http.MethodGet)
	assertEqual(t, entry["path"], "/stations")
	assertEqual(t, entry["route"], "/stations")
	assertEqual(t, entry["query"], "limit=10")
	assertEqual(t, entry["status"], float64(http.StatusCreated))
	assertEqual(t, entry["user_id"], float64(42))
	assertEqual(t, entry["user_agent"], "test-agent")
	if _, ok := entry["latency_ms"]; !ok {
		t.Fatal("missing latency_ms")
	}
}

func TestLoggerStructuredUsesErrorLevelForServerErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	restoreLogger(&logs, t)

	r := gin.New()
	r.Use(RequestID(), LoggerStructured())
	r.GET("/stations", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodGet, "/stations", nil)
	req.Header.Set(HeaderRequestID, "req-2")
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	entry := firstLogEntry(t, logs.String())
	assertEqual(t, entry["msg"], "http_request")
	assertEqual(t, entry["level"], "ERROR")
	assertEqual(t, entry["request_id"], "req-2")
	assertEqual(t, entry["status"], float64(http.StatusInternalServerError))
}

func TestRecoverLogsPanicWithRequestMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	restoreLogger(&logs, t)

	r := gin.New()
	r.Use(RequestID(), Recover())
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set(HeaderRequestID, "req-3")
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", res.Code)
	}

	entry := firstLogEntry(t, logs.String())
	assertEqual(t, entry["msg"], "panic")
	assertEqual(t, entry["level"], "ERROR")
	assertEqual(t, entry["request_id"], "req-3")
	assertEqual(t, entry["method"], http.MethodGet)
	assertEqual(t, entry["path"], "/panic")
	assertEqual(t, entry["route"], "/panic")
	assertEqual(t, entry["panic_type"], "string")
	assertEqual(t, entry["panic_message"], "boom")
	if _, ok := entry["panic"]; ok {
		t.Fatal("raw panic field should not be logged")
	}
	if stack, ok := entry["stack"].(string); !ok || !strings.Contains(stack, "runtime/debug.Stack") {
		t.Fatalf("missing stack: %#v", entry["stack"])
	}
}

func restoreLogger(buf *bytes.Buffer, t *testing.T) {
	t.Helper()
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo})))
	t.Cleanup(func() {
		slog.SetDefault(prev)
	})
}

func firstLogEntry(t *testing.T, raw string) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("missing log entry: %q", raw)
	}
	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("decode log entry: %v\n%s", err, lines[0])
	}
	return entry
}

func assertEqual(t *testing.T, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
