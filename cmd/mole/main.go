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

func connectToSignalingServer(code string, role string, filePath string) {
	serverURL := "wss://moledrop-production.up.railway.app/ws"

	//we connect to the signaling server
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatal("Can't connect to server", err)
	}

	//we send the code to the server to join the room
	err = conn.WriteMessage(websocket.TextMessage, []byte(code))
	if err != nil {
		log.Println("Write error:", err)
		return
	}

	if role == "sender" {
		fmt.Println("Waiting for receiver to connect...")
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Fatal("Lost connection to signaling server:", err)
			}
			switch string(msg) {
			case "waiting":
				// still waiting, loop again
			case "peer-joined":
				fmt.Println("Receiver connected! Starting transfer...")
				p2p.StartSender(conn, filePath)
				return
			case "room-full":
				log.Fatal("Room is full")
			}
		}
	} else if role == "receiver" {
		fmt.Println("Connecting to sender...")
		p2p.StartReceiver(conn)
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
			connectToSignalingServer(code, "sender", file)
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
			connectToSignalingServer(code, "receiver", "")
		},
	}

	//adding the commands to the root command
	rootCmd.AddCommand(sendCmd, receiveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
