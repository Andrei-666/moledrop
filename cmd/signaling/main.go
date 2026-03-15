package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// this is a simple signaling server that will be used to exchange the connection information between the sender and the receiver
//the signaling server role is to create rooms where senders and receivers can join using a unique code
// after that the connection is established between the sender and the receiver and the room is closed when the connection is closed
//that way the signaling server doesnt have any acces to the data being transferred and is only used for the initial connection setup

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Room struct {
	Sender   *websocket.Conn
	Receiver *websocket.Conn
}

// we use a map to store the connections for each room (code)
var rooms = make(map[string]*Room)

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
	room, exists := rooms[roomCode]
	if !exists {
		//if the room doesn't exist, we create it assign Sender role to the current connection

		rooms[roomCode] = &Room{Sender: conn}
		room = rooms[roomCode]
		fmt.Printf("Sender created room: %s\n", roomCode)
	} else {
		//if the room exists, we assign Receiver role to the current connection and notify the sender that the tunnel is ready
		room.Receiver = conn
		fmt.Printf("Receiver joined room: %s\n", roomCode)

		room.Sender.WriteMessage(websocket.TextMessage, []byte("peer-joined"))
	}

	//we release the mutex lock
	mutex.Unlock()

	//every message received from the client will be forwarded to the other peer in the room
	for {
		msgType, payload, err := conn.ReadMessage()
		if err != nil {
			break
		}

		mutex.Lock()
		if room.Sender == conn && room.Receiver != nil {
			//if the sender sends a message and the receiver is connected, we forward the message to the receiver
			room.Receiver.WriteMessage(msgType, payload)
		} else if room.Receiver == conn && room.Sender != nil {
			//if the receiver sends a message and the sender is connected, we forward the message to the sender
			room.Sender.WriteMessage(msgType, payload)
		}
		mutex.Unlock()
	}
	//if the connection is closed, we clean up the room
	mutex.Lock()
	delete(rooms, roomCode)
	mutex.Unlock()
	fmt.Printf("Room closed: %s\n", roomCode)
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	fmt.Println("Signaling server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
