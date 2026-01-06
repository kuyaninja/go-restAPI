package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestMaskSensitiveBodyFields_MasksSensitiveValues(t *testing.T) {
	body := []byte(`{"email":"user@example.com","password":"secret","Token":"abc","nested":{"Api_Key":"hidden"}}`)
	masked := maskSensitiveBodyFields(body)
	if string(masked) == string(body) {
		t.Fatalf("expected sensitive fields to be masked")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(masked, &payload); err != nil {
		t.Fatalf("marshal result invalid JSON: %v", err)
	}

	if got := payload["password"]; got != "*****" {
		t.Fatalf("password not masked, got %v", got)
	}
	if got := payload["Token"]; got != "*****" {
		t.Fatalf("token not masked, got %v", got)
	}

	nested := payload["nested"].(map[string]interface{})
	if got := nested["Api_Key"]; got != "*****" {
		t.Fatalf("nested API key not masked, got %v", got)
	}
}

func TestMaskSensitiveBodyFields_LeavesNonJSON(t *testing.T) {
	body := []byte("not json content")
	masked := maskSensitiveBodyFields(body)
	if string(masked) != string(body) {
		t.Fatalf("expected non JSON body to remain untouched")
	}
}

func TestMaskSensitiveBodyFields_MasksNestedCollections(t *testing.T) {
	body := []byte(`{"list":[{"password":"secret"}],"data":{"tokens":[{"token":"abc"}]}}`)
	masked := maskSensitiveBodyFields(body)
	var payload map[string]interface{}
	if err := json.Unmarshal(masked, &payload); err != nil {
		t.Fatalf("invalid JSON result: %v", err)
	}

	list := payload["list"].([]interface{})
	if got := list[0].(map[string]interface{})["password"]; got != "*****" {
		t.Fatalf("nested password not masked, got %v", got)
	}

	data := payload["data"].(map[string]interface{})
	tokens := data["tokens"].([]interface{})
	if got := tokens[0].(map[string]interface{})["token"]; got != "*****" {
		t.Fatalf("token inside nested slice not masked, got %v", got)
	}
}

func TestMaskSensitiveBodyFields_NoSensitiveFieldsReturnsOriginal(t *testing.T) {
	body := []byte(`{"name":"john"}`)
	masked := maskSensitiveBodyFields(body)
	if string(masked) != string(body) {
		t.Fatalf("expected untouched body when no sensitive fields present")
	}
}

func TestTruncateBody_ReturnsSanitizedJSON(t *testing.T) {
	body := []byte(`{"password":"secret","token":"abc"}`)
	got := truncateBody(body)
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(got), &payload); err != nil {
		t.Fatalf("truncateBody returned invalid JSON: %v", err)
	}

	if payload["password"] != "*****" || payload["token"] != "*****" {
		t.Fatalf("expected sensitive values masked, got %v", payload)
	}
}

func TestTruncateBody_TruncatesLongPayload(t *testing.T) {
	longBody := bytes.Repeat([]byte("a"), maxLoggedBodyBytes+10)
	got := truncateBody(longBody)
	suffix := "...<truncated>"
	if !strings.HasSuffix(got, suffix) {
		t.Fatalf("expected truncated suffix, got %q", got[len(got)-20:])
	}
	if len(got) != maxLoggedBodyBytes+len(suffix) {
		t.Fatalf("unexpected truncated length: got %d want %d", len(got), maxLoggedBodyBytes+len(suffix))
	}
}
