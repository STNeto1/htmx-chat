package main

import (
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/storage/sqlite3/v2"
	"github.com/gofiber/template/html/v2"
)

var store *session.Store

func main() {
	store = createStore()

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(logger.New())

	app.Get("/", HandleIndex)

	app.Get("/room/:room", HandleRoom)
	app.Post("/room", HandleCreateRoom)

	app.Post("/signup", HandleSignup)
	app.Post("/signout", HandleSignout)

	app.Use(HandleWsMiddleware).Get("/ws/:room", websocket.New(HandleMessage, websocket.Config{}))

	go RunHub()

	app.Listen(":3000")
}

func createStore() *session.Store {
	sqliteStorage := sqlite3.New(sqlite3.Config{
		Database:        "./fiber.sqlite3",
		Table:           "fiber_storage",
		Reset:           false,
		GCInterval:      10 * time.Second,
		MaxOpenConns:    100,
		MaxIdleConns:    100,
		ConnMaxLifetime: 1 * time.Second,
	})

	return session.New(session.Config{
		Expiration:   1 * time.Hour,
		Storage:      sqliteStorage,
		KeyLookup:    "cookie:session",
		KeyGenerator: utils.UUIDv4,
	})
}
