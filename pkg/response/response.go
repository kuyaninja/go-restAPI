package response

import "github.com/gofiber/fiber/v2"

// APIResponse standardizes HTTP responses sent back to the client.
type APIResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// Success builds a successful JSON response.
func Success(c *fiber.Ctx, status int, message string, data interface{}) error {
	resp := APIResponse{
		Message: message,
		Data:    data,
	}

	return c.Status(status).JSON(resp)
}

// Failure builds an error JSON response.
func Failure(c *fiber.Ctx, status int, message string, err interface{}) error {
	resp := APIResponse{
		Message: message,
		Error:   err,
	}

	return c.Status(status).JSON(resp)
}
