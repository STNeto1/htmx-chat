package main

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

var engine *html.Engine

func main() {
	engine = html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", HandleIndex)
	app.Get("/room/:room", HandleRoom)
	app.Get("/ws/:room", websocket.New(HandleMessage))

	go RunHub()

	app.Listen(":3000")
}
