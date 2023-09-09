package main

import (
	"errors"
	"fmt"
)

type Room struct {
	ID        int    `db:"id"`
	Name      string `db:"name"`
	CreatedAt string `db:"created_at"`
}

type Message struct {
	ID        int    `db:"id"`
	Message   string `db:"message"`
	User      string `db:"user"`
	RoomID    int    `db:"room_id" json:"-"`
	CreatedAt string `db:"created_at"`
}

func GetRooms() ([]Room, error) {
	var rooms []Room
	err := db.Select(&rooms, "SELECT * FROM rooms")
	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func CreateRoom(name string) error {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM rooms WHERE name = ?", name)
	if err != nil {
		return errors.New("Error checking if room exists")
	}

	if count > 0 {
		return errors.New("Room already exists")
	}

	_, err = db.Exec("INSERT INTO rooms (name) VALUES (?)", name)
	if err != nil {
		return err
	}

	return nil
}

func GetRoom(roomName string) (Room, error) {
	var room Room
	err := db.Get(&room, "SELECT * FROM rooms WHERE name = ?", roomName)
	if err != nil {
		fmt.Println("Error getting room", err)
		return room, err
	}

	return room, nil
}

func CreateMessage(user, message, roomName string) (Message, error) {
	room, err := GetRoom(roomName)
	if err != nil {
		return Message{}, err
	}

	_, err = db.Exec("INSERT INTO messages (user, message, room_id) VALUES (?, ?, ?)", user, message, room.ID)
	if err != nil {
		return Message{}, err
	}

	return Message{
		User:    user,
		Message: message,
		RoomID:  room.ID,
	}, nil
}

func GetRoomMessages(rID int) ([]Message, error) {
	var messages []Message
	err := db.Select(&messages, "SELECT * FROM messages WHERE room_id = ?", rID)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
