package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// this is a simple signaling server that will be used to exchange the connection information between the sender and the receiver
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// we use a map to store the connections for each room (code)
var rooms = make(map[string]*websocket.Conn)

// we use a mutex to protect the rooms map from concurrent access
var mutex sync.Mutex

func wsHandler(w http.ResponseWriter, r *http.Request) {
	//we accept the connection and upgrade it to a websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	//we read the first message from the client which should contain the room code
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Read error:", err)
		return
	}
	roomCode := string(msg)

	//we verify if someone is already in the room, if not we create a new room
	mutex.Lock()
	//if the first user is already in the room, we add the second user to the room and connect them
	if peer, exists := rooms[roomCode]; exists {

		fmt.Printf("Peer joined room: %s\n", roomCode)

		//we tell both peers they are in the same room
		peer.WriteMessage(websocket.TextMessage, []byte("ready"))
		conn.WriteMessage(websocket.TextMessage, []byte("ready"))

		//we delete the room and connect the two peers directly
		delete(rooms, roomCode)

	} else {
		//otherwise we create a new room and wait for the second user to join
		rooms[roomCode] = conn
		fmt.Printf("Room created, waiting for the second user: %s\n", roomCode)
	}
	//we release the mutex lock
	mutex.Unlock()

	//we keep the connection open until the client closes it
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}

	}

}

func main() {
	http.HandleFunc("/ws", wsHandler)
	fmt.Println("Signaling server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
