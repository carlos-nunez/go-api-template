package ws

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	subscribeMessageType = "subscribe"
)

type subscribeMessage struct {
	Type   string `json:"type"`
	RoomID string `json:"roomID"`
	Token  string `json:"token"` // Add this field
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by returning true
		return true
	},
}

type Client struct {
	hub           *Hub
	roomID        roomKey
	conn          *websocket.Conn
	send          chan []byte
	authenticated bool
}

func authenticate(apiToken string) bool {
	uuid := os.Getenv("uuid")
	isAuthenticated := false
	if uuid == "api" {
		_, err := api.GetUserByWSToken(apiToken)
		if err == nil {
			isAuthenticated = true
		}
	} else {
		// fetch websocket server by uuid
		server, err := api.GetWSServerByUUID(uuid)
		if err != nil {
			return false
		}
		if server.ApiToken == apiToken {
			isAuthenticated = true
		}
	}
	//otherwise use the api_key env variable

	return isAuthenticated

}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if c.authenticated {
			var subMsg subscribeMessage
			if err := json.Unmarshal(message, &subMsg); err == nil && subMsg.Type == subscribeMessageType {
				// Authenticate the subscription message using the provided token
				if authenticate(subMsg.Token) {
					// Client is allowed to join the room
					c.roomID = roomKey{Name: subMsg.RoomID + subMsg.Token, Token: subMsg.Token}
					c.hub.register <- c
				}
			} else {
				message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
				c.hub.broadcast <- messagePayload{roomID: c.roomID, data: message}
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Extract the API token, e.g., using a query parameter "token"
	apiToken := r.URL.Query().Get("token")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), authenticated: false}

	// Authenticate the client using the API token
	if authenticate(apiToken) {
		client.authenticated = true
		client.hub.authenticatedClients[client] = struct{}{}
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
