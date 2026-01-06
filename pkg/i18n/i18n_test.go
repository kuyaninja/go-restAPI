package i18n

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestLoadTranslatorLoadsMessages(t *testing.T) {
	dir := t.TempDir()
	writeLocale(t, dir, "messages.en.yaml", "messages:\n  hello: Hello\n")
	writeLocale(t, dir, "messages.es.yaml", "messages:\n  hello: Hola\n")

	translator, err := LoadTranslator(dir, "en")
	if err != nil {
		t.Fatalf("LoadTranslator returned error: %v", err)
	}
	if got := translator.T("en", "hello"); got != "Hello" {
		t.Fatalf("expected english translation, got %q", got)
	}
	if got := translator.T("es", "hello"); got != "Hola" {
		t.Fatalf("expected spanish translation, got %q", got)
	}
}

func TestTranslatorTFallbacks(t *testing.T) {
	translator := &Translator{
		messages: map[string]map[string]string{
			"en": {"hello": "Hello"},
		},
		defaultLocale: "en",
	}

	if got := translator.T("fr", "hello"); got != "Hello" {
		t.Fatalf("expected default locale fallback, got %q", got)
	}
	if got := translator.T("fr", "missing"); got != "missing" {
		t.Fatalf("expected key fallback, got %q", got)
	}
	if got := (*Translator)(nil).T("en", "hello"); got != "hello" {
		t.Fatalf("nil translator should return key, got %q", got)
	}
}

func TestTranslatorMiddlewareStoresLocale(t *testing.T) {
	translator := &Translator{
		messages: map[string]map[string]string{
			"en": {"hello": "Hello"},
			"id": {"hello": "Halo"},
		},
		defaultLocale: "id",
	}

	app := fiber.New()
	app.Use(translator.Middleware())
	app.Get("/", func(c *fiber.Ctx) error {
		if locale := translator.LocaleFromContext(c); locale != "en" {
			t.Fatalf("expected locale en from context, got %q", locale)
		}
		return nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "fr-FR, en;q=0.9")
	if resp, err := app.Test(req); err != nil || resp.StatusCode != 200 {
		t.Fatalf("unexpected response: %v status %d", err, resp.StatusCode)
	}
}

func TestLocaleFromContextFallsBackToDefault(t *testing.T) {
	translator := &Translator{defaultLocale: "en"}
	if locale := translator.LocaleFromContext(nil); locale != "en" {
		t.Fatalf("expected default locale, got %q", locale)
	}
}

func TestTranslatorMiddlewarePrefersLocaleHeader(t *testing.T) {
	translator := &Translator{
		messages: map[string]map[string]string{
			"en": {"hello": "Hello"},
			"id": {"hello": "Halo"},
		},
		defaultLocale: "id",
	}

	app := fiber.New()
	app.Use(translator.Middleware())
	var resolved string
	app.Get("/", func(c *fiber.Ctx) error {
		resolved = translator.LocaleFromContext(c)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(localeHeader, "en")
	req.Header.Set("Accept-Language", "id-ID")
	if resp, err := app.Test(req); err != nil || resp.StatusCode != fiber.StatusOK {
		t.Fatalf("unexpected response: %v status %d", err, resp.StatusCode)
	}
	if resolved != "en" {
		t.Fatalf("expected header locale en, got %q", resolved)
	}
}

func TestTranslatorMiddlewareFallsBackWhenHeaderUnknown(t *testing.T) {
	translator := &Translator{
		messages: map[string]map[string]string{
			"en": {"hello": "Hello"},
			"id": {"hello": "Halo"},
		},
		defaultLocale: "id",
	}

	app := fiber.New()
	app.Use(translator.Middleware())
	var resolved string
	app.Get("/", func(c *fiber.Ctx) error {
		resolved = translator.LocaleFromContext(c)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(localeHeader, "fr")
	req.Header.Set("Accept-Language", "en-US")
	if resp, err := app.Test(req); err != nil || resp.StatusCode != fiber.StatusOK {
		t.Fatalf("unexpected response: %v status %d", err, resp.StatusCode)
	}
	if resolved != "en" {
		t.Fatalf("expected Accept-Language fallback en, got %q", resolved)
	}
}

func TestLocaleFromContextIgnoresNonStringValues(t *testing.T) {
	translator := &Translator{defaultLocale: "en"}
	app := fiber.New()
	var resolved string
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("locale", 123)
		resolved = translator.LocaleFromContext(c)
		return c.SendStatus(fiber.StatusOK)
	})
	if resp, err := app.Test(httptest.NewRequest("GET", "/", nil)); err != nil || resp.StatusCode != fiber.StatusOK {
		t.Fatalf("unexpected response: %v status %d", err, resp.StatusCode)
	}
	if resolved != "en" {
		t.Fatalf("expected fallback to default locale, got %q", resolved)
	}
}

func TestTranslatorMiddlewareNilInstance(t *testing.T) {
	app := fiber.New()
	var translator *Translator
	app.Use(translator.Middleware())
	hit := false
	app.Get("/", func(*fiber.Ctx) error { hit = true; return nil })
	if resp, err := app.Test(httptest.NewRequest("GET", "/", nil)); err != nil || resp.StatusCode != 200 {
		t.Fatalf("unexpected response: %v status %d", err, resp.StatusCode)
	}
	if !hit {
		t.Fatalf("expected handler to run")
	}
}

func TestDefaultLocaleNilTranslator(t *testing.T) {
	var translator *Translator
	if translator.DefaultLocale() != "" {
		t.Fatalf("expected empty default locale for nil translator")
	}
}

func TestDefaultLocaleEmptyValue(t *testing.T) {
	translator := &Translator{}
	if translator.DefaultLocale() != "" {
		t.Fatalf("expected empty string when default locale unset")
	}
}

func TestResolveLocaleFallbacks(t *testing.T) {
	translator := &Translator{
		messages:      map[string]map[string]string{"en": {}, "fr": {}},
		defaultLocale: "fr",
	}
	if locale := translator.resolveLocale(" "); locale != "fr" {
		t.Fatalf("expected default locale on blank header")
	}
	if locale := translator.resolveLocale(";; , en;q=0.8"); locale != "en" {
		t.Fatalf("expected en locale, got %q", locale)
	}
	if locale := translator.resolveLocale("de-DE"); locale != "fr" {
		t.Fatalf("expected fallback to default when locale missing, got %q", locale)
	}
}

func TestLoadTranslatorSkipsInvalidFiles(t *testing.T) {
	dir := t.TempDir()
	writeLocale(t, dir, "custom.en.yml", "messages:\n  hello: Hi\n")
	writeLocale(t, dir, "messages..yaml", "messages:\n")
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("ignored"), 0o644); err != nil {
		t.Fatalf("failed to write text file: %v", err)
	}
	subDir := filepath.Join(dir, "nested")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("failed to make nested dir: %v", err)
	}
	writeLocale(t, subDir, "messages.es.yaml", "messages:\n  hello: Hola\n")

	translator, err := LoadTranslator(dir, "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if translator.T("en", "hello") != "Hi" {
		t.Fatalf("expected custom yaml to load")
	}
	if translator.T("es", "hello") != "Hola" {
		t.Fatalf("expected nested locale to load")
	}
}

func TestLoadTranslatorReturnsErrorOnInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "messages.en.yaml"), []byte("invalid: ["), 0o644); err != nil {
		t.Fatalf("failed to write invalid file: %v", err)
	}
	if _, err := LoadTranslator(dir, "en"); err == nil {
		t.Fatalf("expected error for invalid yaml")
	}
}

func writeLocale(t *testing.T, dir, name, body string) {
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("failed to write locale file: %v", err)
	}
}
