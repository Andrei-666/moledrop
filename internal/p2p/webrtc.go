package p2p

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type SignalMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func SetupWebRTCConnection() *webrtc.PeerConnection {
	//we configure STUN servers
	//STUN servers are used to discover the public IP address and port of the peers behind NAT
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
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

func StartSender(ws *websocket.Conn) {

	//we create a new WebRTC peer connection
	pc := SetupWebRTCConnection()
	defer pc.Close()

	//we create a data channel for file transfer
	dataChannel, err := pc.CreateDataChannel("file-transfer", nil)
	if err != nil {
		log.Fatal("Error creating data channel:", err)
	}

	//we set up the data channel event handlers
	dataChannel.OnOpen(func() {
		fmt.Println("Data channel is open! Starting file transfer")
		//TODO: implement the file transfer logic here
	})

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		log.Fatal("Error creating offer:", err)
	}

	pc.SetLocalDescription(offer)

	//wait for gathering of ICE candidates to complete before sending the offer to the signaling server
	//ICE candidates are the possible network paths that can be used to establish the connection between the peers
	gatherComplete := webrtc.GatheringCompletePromise(pc)
	<-gatherComplete

	//we send the offer to the signaling server to be forwarded to the receiver
	offerBytes, _ := json.Marshal(pc.LocalDescription())
	//we create a signal message with the offer and send it to the signaling server
	msg := SignalMessage{
		Type:    "offer",
		Payload: string(offerBytes),
	}

	//we send the signal message to the signaling server
	ws.WriteJSON(msg)

	var answerMsg SignalMessage
	err = ws.ReadJSON(&answerMsg)
	if err != nil {
		log.Println("Error reading answer from signaling server:", err)
	}
	if answerMsg.Type == "answer" {
		//we unmarshal the answer and set it as the remote description of the peer connection
		var answer webrtc.SessionDescription
		json.Unmarshal([]byte(answerMsg.Payload), &answer)

		pc.SetRemoteDescription(answer)
	}

	//keep the connection alive until the file transfer is complete
	select {}

}

func StartReceiver(ws *websocket.Conn) {
	//we create a new WebRTC peer connection
	pc := SetupWebRTCConnection()
	defer pc.Close()

	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("Data channel created: %s\n", d.Label())

		d.OnOpen(func() {
			fmt.Println("Data channel is open! Ready to receive files")
		})
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Received message: %s\n", string(msg.Data))
		})
	})

	var offerMsg SignalMessage
	fmt.Println("Waiting for offer from sender")
	err := ws.ReadJSON(&offerMsg)
	if err != nil {
		log.Println("Error reading offer from signaling server:", err)
	}

	if offerMsg.Type == "offer" {
		fmt.Println("Received an offer, processing...")

		//we unmarshal the offer and set it as the remote description of the peer connection
		var offer webrtc.SessionDescription
		json.Unmarshal([]byte(offerMsg.Payload), &offer)

		//we set the description as the remote description of the peer connection, this will trigger the WebRTC connection process and generate an answer
		pc.SetRemoteDescription(offer)

		answer, err := pc.CreateAnswer(nil)
		if err != nil {
			log.Fatal(err)
		}
		pc.SetLocalDescription(answer)
		//wait for gathering of ICE candidates to complete before sending the answer to the signaling server
		gatherComplete := webrtc.GatheringCompletePromise(pc)
		<-gatherComplete

		//send the answer back to the signaling server to be forwarded to the sender
		answerBytes, _ := json.Marshal(pc.LocalDescription())
		msg := SignalMessage{
			Type:    "answer",
			Payload: string(answerBytes),
		}

		fmt.Println("Send the response to the peer")
		ws.WriteJSON(msg)
	}

	select {}

}
