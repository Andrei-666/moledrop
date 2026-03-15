package p2p

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type SignalMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// progress bar function to display the progress of the file transfer in the console
func printProgressBar(current, total int64, prefix string, startTime time.Time) {
	if total == 0 {
		return
	}
	percent := float64(current) / float64(total) * 100

	elapsed := time.Since(startTime).Seconds()
	var speed float64
	if elapsed > 0 {
		speed = float64(current) / elapsed / 1024 / 1024 // speed in MB/s
	}

	barLength := 40
	filled := int(float64(barLength) * float64(current) / float64(total))
	if filled > barLength {
		filled = barLength
	}

	bar := strings.Repeat("=", filled) + strings.Repeat("-", barLength-filled)
	fmt.Printf("\r%s [%s] %.2f%% | %.2f MB/s  ", prefix, bar, percent, speed)
}

func SetupWebRTCConnection() *webrtc.PeerConnection {

	//Architectural overview of the sender function:
	// file -> read in chunks -> data channel -> WebRTC connection -> signaling server -> receiver

	s := webrtc.SettingEngine{}
	s.DetachDataChannels()
	s.SetSCTPMaxReceiveBufferSize(16 * 1024 * 1024)

	api := webrtc.NewAPI(webrtc.WithSettingEngine(s))

	//we configure STUN servers
	//STUN servers are used to discover the public IP address and port of the peers behind NAT

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
			{
				URLs: []string{"stun:stun.1.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		log.Fatal("Error creating peer connection:", err)
	}

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Connection state changed: %s\n", s.String())

		//verify if the peers are connected, if they are we can start the file transfer

		if s == webrtc.PeerConnectionStateConnected {
			fmt.Println("Peers are connected! Ready to transfer files")
		}

	})

	return peerConnection

}

func StartSender(ws *websocket.Conn, filePath string) {

	//we create a new WebRTC peer connection
	pc := SetupWebRTCConnection()
	defer pc.Close()

	//we create a data channel for file transfer
	ordered := true
	dataChannel, _ := pc.CreateDataChannel("file-transfer", &webrtc.DataChannelInit{
		Ordered: &ordered,
	})
	done := make(chan bool)

	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		if string(msg.Data) == "DONE" {
			fmt.Println("File transfer complete! Closing connection.")
			done <- true
		}
	})

	//we set up the data channel event handlers
	dataChannel.OnOpen(func() {
		fmt.Println("Data channel is open! Starting file transfer")

		//we detach the data channel to get a raw channel that we can use to send large files without worrying about the internal buffering of the data channel
		raw, err := dataChannel.Detach()
		if err != nil {
			fmt.Println("Error detaching data channel:", err)
			return
		}
		defer raw.Close()

		//we read the file to be sent and send it over the data channel
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}

		actualFilePath := filePath
		isTempZip := false

		if fileInfo.IsDir() {
			fmt.Println("The specified path is a directory. Zipping it before transfer.")
			actualFilePath = filepath.Join(os.TempDir(), fileInfo.Name()+".zip")
			err := zipDirectory(filePath, actualFilePath)
			if err != nil {
				fmt.Println("Error occurred while zipping:", err)
				return
			}
			isTempZip = true
			fileInfo, _ = os.Stat(actualFilePath) // Actualizăm detaliile pentru fișierul zip creat
		}

		//we read the file in chunks and send each chunk over the data channel
		file, err := os.Open(actualFilePath)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer file.Close()

		if isTempZip {
			defer os.Remove(actualFilePath)
		}

		totalSize := fileInfo.Size()

		metadata := fmt.Sprintf("METADATA:%s:%d", fileInfo.Name(), totalSize)
		raw.Write([]byte(metadata))

		startTime := time.Now()
		buffer := make([]byte, 1024*16) //16KB buffer

		var sentSize int64
		var lastUpdate int64

		hasher := sha256.New()

		for {
			for dataChannel.BufferedAmount() > 2*1024*1024 {
				time.Sleep(2 * time.Millisecond)
			}
			n, err := file.Read(buffer)
			if n > 0 {
				hasher.Write(buffer[:n])
				_, writeErr := raw.Write(buffer[:n])
				if writeErr != nil {
					fmt.Printf("\nSending error: %v\n", writeErr)
					return
				}

				sentSize += int64(n)
				if sentSize-lastUpdate > 1024*1024 || sentSize == totalSize {
					printProgressBar(sentSize, totalSize, "Progress", startTime)
					lastUpdate = sentSize
				}
			}
			if err != nil {
				break //end of file or error
			}
		}

		for dataChannel.BufferedAmount() > 0 {
			time.Sleep(10 * time.Millisecond)
		}

		hashString := hex.EncodeToString(hasher.Sum(nil))

		fmt.Printf("\nTransfer complete! Hash: %s\n", hashString)
		done <- true //close the connection after the transfer is complete
	})

	//we create an offer for the WebRTC connection and set it as the local description
	offer, _ := pc.CreateOffer(nil)
	pc.SetLocalDescription(offer)
	<-webrtc.GatheringCompletePromise(pc)

	offerBytes, _ := json.Marshal(pc.LocalDescription())
	ws.WriteJSON(SignalMessage{Type: "offer", Payload: string(offerBytes)})

	var answerMsg SignalMessage
	ws.ReadJSON(&answerMsg)
	if answerMsg.Type == "answer" {
		var answer webrtc.SessionDescription
		json.Unmarshal([]byte(answerMsg.Payload), &answer)
		pc.SetRemoteDescription(answer)
	}

	<-done
}

func StartReceiver(ws *websocket.Conn) {
	pc := SetupWebRTCConnection()
	defer pc.Close()

	done := make(chan bool)

	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("Data channel created: %s\n", d.Label())

		d.OnOpen(func() {
			fmt.Println("Data channel is open! Ready to receive files")

			raw, err := d.Detach()
			if err != nil {
				fmt.Println("Error detaching data channel:", err)
				return
			}
			defer raw.Close()

			buffer := make([]byte, 16*1024)

			//wait for the metadata message first to get the file name and size
			n, err := raw.Read(buffer)
			if err != nil {
				fmt.Println("Error reading metadata:", err)
				return
			}

			payload := string(buffer[:n])
			if !strings.HasPrefix(payload, "METADATA:") {
				fmt.Println("Error: First message was not metadata!")
				return
			}

			parts := strings.Split(payload, ":")
			name := parts[1]
			totalSize, _ := strconv.ParseInt(parts[2], 10, 64)

			newFile, _ := os.Create(name)
			defer newFile.Close()

			bufferedWriter := bufio.NewWriterSize(newFile, 1024*1024)
			defer bufferedWriter.Flush()

			fmt.Printf("Receiving: %s (%d MB)\n", name, totalSize/1024/1024)

			startTime := time.Now()
			var receivedSize int64
			var lastUpdate int64

			hasher := sha256.New()

			// we read the file data in chunks and write it to the new file until the transfer is complete
			for {
				n, err := raw.Read(buffer)
				if n > 0 {
					bufferedWriter.Write(buffer[:n])
					hasher.Write(buffer[:n])
					receivedSize += int64(n)

					if receivedSize-lastUpdate > 1024*1024 || receivedSize == totalSize {
						printProgressBar(receivedSize, totalSize, "Downloading", startTime)
						lastUpdate = receivedSize
					}
				}

				if err != nil {
					bufferedWriter.Flush()
					hashString := hex.EncodeToString(hasher.Sum(nil))
					//close the file and exit the loop when the transfer is complete or if there is an error
					fmt.Println("\nFile saved successfully! Closing connection. Hash:", hashString)
					break
				}
			}
			done <- true
		})
	})

	//we wait for the offer from the sender, set it as the remote description and create an answer to establish the WebRTC connection
	var offerMsg SignalMessage
	fmt.Println("Waiting for offer from sender")
	ws.ReadJSON(&offerMsg)

	if offerMsg.Type == "offer" {
		var offer webrtc.SessionDescription
		json.Unmarshal([]byte(offerMsg.Payload), &offer)
		pc.SetRemoteDescription(offer)

		answer, _ := pc.CreateAnswer(nil)
		pc.SetLocalDescription(answer)
		<-webrtc.GatheringCompletePromise(pc)

		answerBytes, _ := json.Marshal(pc.LocalDescription())
		ws.WriteJSON(SignalMessage{Type: "answer", Payload: string(answerBytes)})
	}

	<-done
}
