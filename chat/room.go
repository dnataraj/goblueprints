package main

import (
	"github.com/dnataraj/goblueprints/trace"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type room struct {
	forward chan []byte // holds incoming messages that should be forwarded to other clients
	join    chan *client
	leave   chan *client
	clients map[*client]bool
	tracer  trace.Tracer
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// client is joining
			r.clients[client] = true
			r.tracer.Trace("new client joined")
		case client := <-r.leave:
			// client is leaving
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("client left")
		case msg := <-r.forward:
			r.tracer.Trace("message received: ", string(msg))
			// forward msg to all clients
			for client := range r.clients {
				client.send <- msg
				r.tracer.Trace(" -- sent to client")
			}
		}
	}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}
