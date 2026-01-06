package response

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSuccessWritesResponse(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return Success(c, fiber.StatusCreated, "done", map[string]string{"id": "123"})
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	var payload APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.Message != "done" || payload.Error != nil {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload.Data.(map[string]interface{})["id"].(string) != "123" {
		t.Fatalf("unexpected data: %#v", payload.Data)
	}
}

func TestFailureWritesResponse(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return Failure(c, fiber.StatusBadRequest, "bad", fiber.Map{"code": "invalid"})
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
	var payload APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.Message != "bad" || payload.Data != nil {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload.Error.(map[string]interface{})["code"].(string) != "invalid" {
		t.Fatalf("unexpected error payload: %#v", payload.Error)
	}
}
