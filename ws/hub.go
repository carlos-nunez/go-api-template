package ws

import (
	API "github.com/carlos-nunez/go-api-template/api"
)

type messagePayload struct {
	roomID roomKey
	data   []byte
}

type roomKey struct {
	Name  string
	Token string
}

type Hub struct {
	// Registered clients.
	rooms map[roomKey]map[*Client]struct{}

	authenticatedClients map[*Client]struct{}

	// Inbound messages from the clients.
	broadcast chan messagePayload

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

var api *API.API

func NewHub(apiRef *API.API) *Hub {
	api = apiRef
	return &Hub{
		broadcast:            make(chan messagePayload),
		register:             make(chan *Client),
		unregister:           make(chan *Client),
		rooms:                make(map[roomKey]map[*Client]struct{}),
		authenticatedClients: make(map[*Client]struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			room := h.rooms[client.roomID]
			if room == nil {
				// First client in the room, create a new one
				room = make(map[*Client]struct{})
				h.rooms[client.roomID] = room
			}
			room[client] = struct{}{}
		case client := <-h.unregister:
			room := h.rooms[client.roomID]
			if room != nil {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.send)
					if len(room) == 0 {
						// This was last client in the room, delete the room
						delete(h.rooms, client.roomID)
					}
				}
			}
		case message := <-h.broadcast:
			room := h.rooms[message.roomID]
			if room != nil {
				for client := range room {
					select {
					case client.send <- message.data:
					default:
						close(client.send)
						delete(room, client)
					}
				}
				if len(room) == 0 {
					// The room was emptied while broadcasting to the room.  Delete the room.
					delete(h.rooms, message.roomID)
				}
			}
		}
	}
}
