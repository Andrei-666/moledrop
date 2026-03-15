# MoleDrop

> Ultra-fast, E2E encrypted P2P file transfer utility across CLI and Web.

MoleDrop is a modern file transfer tool that allows you to send files directly from your terminal to any web browser (or another terminal) using WebRTC. No server storage, no relays, just direct RAM-to-RAM transfer.

## ✨ Features
* **True P2P (WebRTC):** Direct connection between peers. The signaling server never touches your files.
* **Cross-Platform:** Works seamlessly between Windows, macOS, Linux (CLI), and any modern Web Browser.
* **End-to-End Encrypted:** Native DTLS/SRTP encryption.
* **Frictionless UX:** Auto-generated human-readable room codes (e.g., `quantum-mole-8831`) and QR codes.

## 🏗️ Architecture
* **CLI / Core:** Go + Pion (WebRTC)
* **Signaling Server:** Go + WebSockets
* **Web Client:** Vanilla JS + WebRTC Streams API

## 🚀 Getting Started (WIP)
*Project is currently in active development. Build instructions will be added in v0.1.*

## 🗺️ Roadmap
- [x] Initial project setup and architecture design
- [ ] Implement secure word-generator for room codes
- [ ] Build WebSocket signaling server
- [ ] Implement CLI sender & receiver (Pion WebRTC)
- [ ] Build Web Receiver (Vanilla JS)
- [ ] Desktop GUI wrapping (Wails)

## 📄 License
This project is licensed under the MIT License.
