package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Andrei-666/moledrop/internal/p2p"
	"github.com/Andrei-666/moledrop/internal/words"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

func connectToSignalingServer(code string, role string) {
	serverURL := "ws://localhost:8080/ws"

	//we connect to the signaling server
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatal("Can't connect to server")
	}

	//we send the code to the server to join the room
	err = conn.WriteMessage(websocket.TextMessage, []byte(code))
	if err != nil {
		log.Println("Write error:", err)
		return
	}

	if role == "sender" {
		//if we are the sender, we wait for the receiver to join the room and send us a "ready" signal before starting the file transfer
		_, msg, err := conn.ReadMessage()
		if err != nil || string(msg) != "ready" {
			log.Println("Error waiting for receiver to join:", err)
			return
		}
		fmt.Println("Receiver joined! Starting file transfer")
		p2p.StartSender(conn)
	} else if role == "receiver" {
		//else if we are the receiver, we wait for the sender to create the room and then start the file transfer
		fmt.Println("Connected, waiting for p2p connection to be established")

		p2p.StartReceiver(conn)
	}

	//we wait for the other peer to join the room
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Read error:", err)
		return
	}

	//if we receive a "ready" signal, we can start the file transfer
	if string(msg) == "ready" {
		fmt.Println("Tunnel is ready! Starting file transfer")

		//TODO: implement the file transfer logic here
	}
}

func main() {
	//starting point of the application
	var rootCmd = &cobra.Command{
		Use:   "moledrop",
		Short: "Moledrop - Ultra-fast P2P file transfer 🦡",
		Long:  `Moledrop is a CLI tool for lightning-fast file sharing between devices in any network`,
	}

	//defining the "send" command
	var sendCmd = &cobra.Command{
		Use:   "send [file]",
		Short: "Generate a unique code for sharing a file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			file := args[0]
			code := words.GenerateCode()
			fmt.Printf("🦡 Mole is digging the tunnel for: %s\n", file)
			fmt.Printf("Share this code with your friend to receive the file: %s\n", code)
			connectToSignalingServer(code, "sender")
		},
	}

	//defining the "receive" command
	var receiveCmd = &cobra.Command{
		Use:   "receive [code]",
		Short: "Receive a file using the unique code",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			code := args[0]
			fmt.Printf("🦡 Mole is searching for the tunnel with the code: %s\n", code)
			connectToSignalingServer(code, "receiver")
		},
	}

	//adding the commands to the root command
	rootCmd.AddCommand(sendCmd, receiveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
