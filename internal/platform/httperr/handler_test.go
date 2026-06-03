package httperr

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"

	"github.com/gin-gonic/gin"
)

func TestHandleLogsUnexpectedErrorAndHidesDetailsFromClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	restoreLogger(&logs, t)

	r := gin.New()
	r.Use(mw.RequestID())
	r.GET("/stations", Wrap(func(c *gin.Context) error {
		return errors.New("database query failed")
	}))

	req := httptest.NewRequest(http.MethodGet, "/stations", nil)
	req.Header.Set(mw.HeaderRequestID, "req-1")
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", res.Code)
	}
	body := responseBody(t, res.Body.String())
	errBody := body["error"].(map[string]any)
	assertEqual(t, errBody["code"], "INTERNAL_ERROR")
	assertEqual(t, errBody["message"], "internal error")
	if strings.Contains(res.Body.String(), "database query failed") {
		t.Fatalf("response leaked internal error: %s", res.Body.String())
	}

	entry := firstLogEntry(t, logs.String())
	assertEqual(t, entry["msg"], "http_error")
	assertEqual(t, entry["level"], "ERROR")
	assertEqual(t, entry["request_id"], "req-1")
	assertEqual(t, entry["method"], http.MethodGet)
	assertEqual(t, entry["path"], "/stations")
	assertEqual(t, entry["route"], "/stations")
	assertEqual(t, entry["status"], float64(http.StatusInternalServerError))
	assertEqual(t, entry["err"], "database query failed")
	assertEqual(t, entry["error_type"], "*errors.errorString")
}

func TestHandleLogsHTTPErrorAsWarning(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	restoreLogger(&logs, t)

	r := gin.New()
	r.Use(mw.RequestID())
	r.GET("/stations", Wrap(func(c *gin.Context) error {
		return BadRequest("invalid limit")
	}))

	req := httptest.NewRequest(http.MethodGet, "/stations", nil)
	req.Header.Set(mw.HeaderRequestID, "req-2")
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", res.Code)
	}

	entry := firstLogEntry(t, logs.String())
	assertEqual(t, entry["msg"], "http_error")
	assertEqual(t, entry["level"], "WARN")
	assertEqual(t, entry["request_id"], "req-2")
	assertEqual(t, entry["status"], float64(http.StatusBadRequest))
	assertEqual(t, entry["err"], "invalid limit")
	assertEqual(t, entry["error_type"], "*httperr.HTTPError")
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

func responseBody(t *testing.T, raw string) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatalf("decode response: %v\n%s", err, raw)
	}
	return body
}

func assertEqual(t *testing.T, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
