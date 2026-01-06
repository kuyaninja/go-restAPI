package controller

import "github.com/gofiber/fiber/v2"

// UserHTTPController exposes the required handler set for routing.
type UserHTTPController interface {
	Version() string
	List(c *fiber.Ctx) error
	Get(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}
