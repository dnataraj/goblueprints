package main

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type client struct {
	socket   *websocket.Conn // websocket for this client
	send     chan *message   // channels on which messages are sent
	room     *room           // the room this client is chatting in
	userData *User
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		var msg *message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			return
		}
		msg.When = time.Now()
		msg.Name = c.userData.Name
		log.Printf("Message : %s\n", msg)
		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		log.Printf("Sending Message : %s\n", msg)
		err := c.socket.WriteJSON(msg)
		if err != nil {
			return
		}
	}
}
