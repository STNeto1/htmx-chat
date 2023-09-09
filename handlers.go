package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type client struct {
	isClosing bool
	mu        sync.Mutex
}

var rooms = make(map[string]map[*websocket.Conn]*client)
var clients = make(map[*websocket.Conn]*client)
var register = make(chan *websocket.Conn)
var broadcast = make(chan string)
var unregister = make(chan *websocket.Conn)
var messages = make(map[string][]string)

func HandleIndex(c *fiber.Ctx) error {
	if len(rooms) == 0 {
		rooms["test"] = make(map[*websocket.Conn]*client)
	}

	keys := make([]string, 0, len(rooms))
	for k := range rooms {
		keys = append(keys, k)
	}

	return c.Render("index", fiber.Map{
		"rooms": keys,
	}, "layouts/main")
}

func HandleRoom(c *fiber.Ctx) error {
	room := c.Params("room")

	if _, ok := rooms[room]; !ok {
		rooms[room] = make(map[*websocket.Conn]*client)
	}

	roomMessages := messages[room]

	return c.Render("room", fiber.Map{
		"room":     room,
		"messages": roomMessages,
	}, "layouts/main")
}

type MessagePayload struct {
	Room    string `json:"room"`
	Message string `json:"message"`
}

func HandleMessage(c *websocket.Conn) {
	// When the function returns, unregister the client and close the connection
	defer func() {
		unregister <- c
		c.Close()
	}()

	// Register the client
	register <- c

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}

			break
		}

		if messageType == websocket.TextMessage {
			var msg MessagePayload
			err := json.Unmarshal(message, &msg)
			if err != nil {
				log.Println("error unmarshalling message:", err)

				// TODO: handle error, but for now just continue
				continue
			}

			// Store the Message
			messages[msg.Room] = append(messages[msg.Room], msg.Message)

			tmpl, err := template.ParseFiles("./views/partials/messages.html")
			if err != nil {
				log.Println("error parsing template:", err)
				continue
			}

			roomMessages := messages[msg.Room]

			var resultHtml bytes.Buffer
			err = tmpl.Execute(&resultHtml, fiber.Map{
				"messages": roomMessages,
			})

			// Broadcast the received message
			broadcast <- string(resultHtml.String())
		} else {
			log.Println("websocket message received of type", messageType)
		}
	}
}

func RunHub() {
	for {
		select {
		case connection := <-register:
			clients[connection] = &client{}
			log.Println("connection registered")

		case message := <-broadcast:
			// Send the message to all clients
			for connection, c := range clients {
				go func(connection *websocket.Conn, c *client) { // send to each client in parallel so we don't block on a slow client
					c.mu.Lock()
					defer c.mu.Unlock()

					if c.isClosing {
						return
					}

					if err := connection.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
						c.isClosing = true
						log.Println("write error:", err)

						connection.WriteMessage(websocket.CloseMessage, []byte{})
						connection.Close()
						unregister <- connection
					}
				}(connection, c)
			}

		case connection := <-unregister:
			// Remove the client from the hub
			delete(clients, connection)

			log.Println("connection unregistered")
		}
	}
}
