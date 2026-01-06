package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	maxLoggedBodyBytes = 2048
	httpResponseLogKey = "http response"
)

// HTTPLogger logs inbound requests and outbound responses using the shared logger.
func HTTPLogger(log *Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if log == nil {
			return c.Next()
		}

		start := time.Now()
		reqFields := Fields{
			"method":     c.Method(),
			"path":       c.Path(),
			"query":      string(c.Request().URI().QueryString()),
			"remote_ip":  c.IP(),
			"user_agent": c.Get("User-Agent"),
		}

		if body := c.Body(); len(body) > 0 {
			reqFields["body"] = truncateBody(body)
		}

		log.Info(c.UserContext(), "http request", reqFields)

		err := c.Next()

		status := c.Response().StatusCode()
		respFields := Fields{
			"method":      c.Method(),
			"path":        c.Path(),
			"status":      status,
			"duration_ms": float64(time.Since(start).Microseconds()) / 1000,
			"bytes":       len(c.Response().Body()),
		}

		if body := c.Response().Body(); len(body) > 0 {
			respFields["body"] = truncateBody(body)
		}

		if err != nil {
			log.Error(c.UserContext(), httpResponseLogKey, err, respFields)
			return err
		}

		if status >= 500 {
			log.Error(c.UserContext(), httpResponseLogKey, nil, respFields)
		} else {
			log.Info(c.UserContext(), httpResponseLogKey, respFields)
		}

		return nil
	}
}

func truncateBody(body []byte) string {
	sanitized := maskSensitiveBodyFields(body)
	if len(sanitized) <= maxLoggedBodyBytes {
		return string(sanitized)
	}

	var buf bytes.Buffer
	buf.Write(sanitized[:maxLoggedBodyBytes])
	buf.WriteString("...<truncated>")
	return buf.String()
}

var sensitiveBodyKeys = map[string]struct{}{
	"password":      {},
	"token":         {},
	"access_token":  {},
	"refresh_token": {},
	"authorization": {},
	"api_key":       {},
	"secret":        {},
	"client_secret": {},
	"client_token":  {},
	"session_token": {},
}

func maskSensitiveBodyFields(body []byte) []byte {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return body
	}

	var payload interface{}
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		return body
	}

	if !scrubSensitiveFields(payload) {
		return body
	}

	masked, err := json.Marshal(payload)
	if err != nil {
		return body
	}

	return masked
}

func scrubSensitiveFields(value interface{}) bool {
	switch v := value.(type) {
	case map[string]interface{}:
		changed := false
		for key, val := range v {
			if _, ok := sensitiveBodyKeys[strings.ToLower(key)]; ok {
				v[key] = "*****"
				changed = true
				continue
			}

			if scrubSensitiveFields(val) {
				changed = true
			}
		}
		return changed
	case []interface{}:
		changed := false
		for i, val := range v {
			if scrubSensitiveFields(val) {
				changed = true
				v[i] = val
			}
		}
		return changed
	default:
		return false
	}
}
