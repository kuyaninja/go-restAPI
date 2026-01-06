package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"skeleton-go/pkg/session"
)

func TestLoggerWritesJSONWithSession(t *testing.T) {
	var buf bytes.Buffer
	log := &Logger{writers: []io.Writer{&buf}}
	ctx := session.ContextWithSession(context.Background(), &session.Session{ID: "abc"})
	log.Info(ctx, "hello", Fields{"foo": "bar"})

	entry := parseLog(t, buf.String())
	if entry["level"] != "info" || entry["message"] != "hello" {
		t.Fatalf("unexpected log entry: %#v", entry)
	}
	if entry["session_id"] != "abc" {
		t.Fatalf("expected session id, got %#v", entry["session_id"])
	}
	fields := entry["fields"].(map[string]interface{})
	if fields["foo"] != "bar" {
		t.Fatalf("fields missing data: %#v", fields)
	}
}

func TestLoggerErrorAddsErrorField(t *testing.T) {
	var buf bytes.Buffer
	log := &Logger{writers: []io.Writer{&buf}}
	log.Error(context.Background(), "oops", errors.New("boom"), nil)

	entry := parseLog(t, buf.String())
	fields := entry["fields"].(map[string]interface{})
	if fields["error"] != "boom" {
		t.Fatalf("expected serialized error, got %#v", fields["error"])
	}
}

func TestLoggerDBWritesMessage(t *testing.T) {
	var buf bytes.Buffer
	log := &Logger{writers: []io.Writer{&buf}}
	log.DB(context.Background(), "query", Fields{"sql": "select"})
	entry := parseLog(t, buf.String())
	if entry["level"] != "db" {
		t.Fatalf("expected db level, got %v", entry["level"])
	}
}

func TestNewCreatesFileWriter(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "app.log")
	log, err := New(Options{File: logFile})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	log.Info(context.Background(), "hello", nil)
	if _, err := os.Stat(logFile); err != nil {
		t.Fatalf("expected log file to be created: %v", err)
	}
}

func TestNewWithoutFile(t *testing.T) {
	if _, err := New(Options{}); err != nil {
		t.Fatalf("New returned error: %v", err)
	}
}

func TestHTTPLoggerLogsAndSanitizes(t *testing.T) {
	var buf bytes.Buffer
	log := &Logger{writers: []io.Writer{&buf}}
	app := fiber.New()
	app.Use(HTTPLogger(log))
	app.Post("/login", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"token": "secret"})
	})

	req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("unexpected test response: %v status %d", err, resp.StatusCode)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(lines))
	}
	request := parseLog(t, lines[0])
	respLog := parseLog(t, lines[1])
	if body := request["fields"].(map[string]interface{})["body"].(string); !strings.Contains(body, "*****") {
		t.Fatalf("expected password masked, got %s", body)
	}
	if body := respLog["fields"].(map[string]interface{})["body"].(string); !strings.Contains(body, "*****") {
		t.Fatalf("expected response token masked, got %s", body)
	}
}

func TestHTTPLoggerWithNilLoggerSkipsLogging(t *testing.T) {
	app := fiber.New()
	app.Use(HTTPLogger(nil))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(204) })
	if resp, err := app.Test(httptest.NewRequest("GET", "/", nil)); err != nil || resp.StatusCode != 204 {
		t.Fatalf("unexpected response: %v status %d", err, resp.StatusCode)
	}
}

func TestHTTPLoggerLogsErrorsAndServerFailures(t *testing.T) {
	var buf bytes.Buffer
	log := &Logger{writers: []io.Writer{&buf}}
	app := fiber.New()
	app.Use(HTTPLogger(log))
	app.Get("/error", func(*fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadGateway, "boom")
	})
	app.Get("/fail", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusServiceUnavailable).SendString("oops")
	})

	if resp, err := app.Test(httptest.NewRequest("GET", "/error", nil)); err != nil || resp.StatusCode != fiber.StatusBadGateway {
		t.Fatalf("unexpected response: %v status %d", err, resp.StatusCode)
	}
	if resp, err := app.Test(httptest.NewRequest("GET", "/fail", nil)); err != nil || resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("unexpected response: %v status %d", err, resp.StatusCode)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 log lines, got %d", len(lines))
	}
	errorEntry := parseLog(t, lines[1])
	if errorEntry["level"] != "error" {
		t.Fatalf("expected error level for exception log")
	}
	if errorEntry["message"] != "http response" {
		t.Fatalf("unexpected log message: %v", errorEntry["message"])
	}
	serverErrorEntry := parseLog(t, lines[3])
	if serverErrorEntry["level"] != "error" {
		t.Fatalf("expected error level for 5xx response")
	}
	if status := serverErrorEntry["fields"].(map[string]interface{})["status"].(float64); status != float64(fiber.StatusServiceUnavailable) {
		t.Fatalf("unexpected logged status: %v", status)
	}
}

func parseLog(t *testing.T, line string) map[string]interface{} {
	t.Helper()
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("failed to parse log line %q: %v", line, err)
	}
	return entry
}
