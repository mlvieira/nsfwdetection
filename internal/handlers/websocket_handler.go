package handlers

import (
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/mlvieira/nsfwdetection/internal/config"
	"github.com/mlvieira/nsfwdetection/internal/models"
	"github.com/mlvieira/nsfwdetection/internal/websockets"
)

var jwtSecretKey = []byte(config.AppConfig.Security.JWTSecretKey)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == config.AppConfig.Server.DomainName
	},
}

// HandleWebSocket handles new WebSocket connections
func HandleWebSocket(hub *websockets.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Sec-WebSocket-Protocol")
		if tokenStr == "" {
			http.Error(w, "Unauthorized: Token is missing", http.StatusUnauthorized)
			return
		}

		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade WebSocket:", err)
			return
		}

		client := &websockets.Client{
			Conn: conn,
			Send: make(chan []byte, 256),
		}

		hub.Register <- client

		go readMessages(client, hub)
		go writeMessages(client)
	}
}

// Reads messages from a client (not used but kept for expansion)
func readMessages(client *websockets.Client, hub *websockets.Hub) {
	defer func() {
		hub.Unregister <- client
		client.Conn.Close()
	}()
	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// Writes messages to the client
func writeMessages(client *websockets.Client) {
	defer client.Conn.Close()
	for message := range client.Send {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}
