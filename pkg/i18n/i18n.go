package i18n

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v3"
)

// Translator fetches localized strings based on keys and locale.
type Translator struct {
	mu            sync.RWMutex
	messages      map[string]map[string]string // locale -> key -> message
	defaultLocale string
}

// LoadTranslator loads all yaml files within the provided directory.
func LoadTranslator(dir, defaultLocale string) (*Translator, error) {
	t := &Translator{messages: make(map[string]map[string]string), defaultLocale: defaultLocale}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		return t.loadLocaleFile(path, d)
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

// T returns the localized message for the key, falling back to the default locale or the key itself.
func (t *Translator) T(locale, key string) string {
	if t == nil {
		return key
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	if locale != "" {
		if msgs, ok := t.messages[locale]; ok {
			if msg := msgs[key]; msg != "" {
				return msg
			}
		}
	}
	if msgs, ok := t.messages[t.defaultLocale]; ok {
		if msg := msgs[key]; msg != "" {
			return msg
		}
	}
	return key
}

// DefaultLocale returns the default locale configured on the translator.
func (t *Translator) DefaultLocale() string {
	if t == nil || t.defaultLocale == "" {
		return ""
	}
	return t.defaultLocale
}

// Middleware stores the resolved locale for each request based on Accept-Language header.
const localeHeader = "X-Locale"

func (t *Translator) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if t == nil {
			return c.Next()
		}
		locale := ""
		if header := c.Get(localeHeader); header != "" {
			if resolved := t.resolveExplicitLocale(header); resolved != "" {
				locale = resolved
			}
		}
		if locale == "" {
			locale = t.resolveLocale(c.Get("Accept-Language"))
		}
		c.Locals("locale", locale)
		return c.Next()
	}
}

// LocaleFromContext reads the locale that Middleware stored, falling back to default.
func (t *Translator) LocaleFromContext(c *fiber.Ctx) string {
	if c == nil {
		return t.DefaultLocale()
	}
	if val, ok := c.Locals("locale").(string); ok && val != "" {
		return val
	}
	return t.DefaultLocale()
}

func (t *Translator) resolveLocale(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return t.DefaultLocale()
	}
	parts := strings.Split(header, ",")
	for _, part := range parts {
		lang := strings.TrimSpace(part)
		if lang == "" {
			continue
		}
		lang = strings.Split(lang, ";")[0]
		lang = strings.ToLower(strings.TrimSpace(lang))
		lang = strings.Split(lang, "-")[0]
		if lang == "" {
			continue
		}
		t.mu.RLock()
		_, ok := t.messages[lang]
		t.mu.RUnlock()
		if ok {
			return lang
		}
	}
	return t.DefaultLocale()
}

func (t *Translator) loadLocaleFile(path string, d fs.DirEntry) error {
	name := d.Name()
	if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
		return nil
	}
	locale := strings.TrimSuffix(name, filepath.Ext(name))
	parts := strings.Split(locale, ".")
	locale = parts[len(parts)-1]
	if locale == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var payload struct {
		Messages map[string]string `yaml:"messages"`
	}
	if err := yaml.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	if len(payload.Messages) == 0 {
		return nil
	}

	t.messages[locale] = payload.Messages
	return nil
}

func (t *Translator) resolveExplicitLocale(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.Split(value, ",")[0]
	value = strings.Split(value, ";")[0]
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.ToLower(value)
	value = strings.Split(value, "-")[0]
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if _, ok := t.messages[value]; ok {
		return value
	}
	return ""
}
