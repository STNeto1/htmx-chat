package main

import "github.com/gofiber/fiber/v2"

func HandleIndex(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{}, "layouts/main")
}
