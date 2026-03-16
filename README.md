# MoleDrop
> Lightweight CLI tool for fast, encrypted, direct device-to-device file transfer.

MoleDrop was built to solve a simple problem: transferring files between two machines quickly, without uploading them to a third-party service. No accounts, no size limits, no cloud storage, just share a code and the transfer starts directly between devices.

---

## Install

**Linux / macOS**
```bash
curl -fsSL https://raw.githubusercontent.com/Andrei-666/moledrop/main/install.sh | sudo bash
```

**Windows** (PowerShell as Administrator)
```powershell
irm https://raw.githubusercontent.com/Andrei-666/moledrop/main/install.ps1 | iex
```

---

## Usage

**Send a file:**
```bash
mole send document.pdf
```
```
Sending: document.pdf
Share this code: crystal-mole-4821
Waiting for receiver to connect...
Receiver connected! Starting transfer...
Sending [====================--------------------] 50.00% | 38.20 MB/s
Done! Hash: 83f4cb...
```

**Receive a file:**
```bash
mole receive crystal-mole-4821
```
```
Connecting with code: crystal-mole-4821
Waiting for sender...
Receiving: document.pdf (124.0 MB)
Receiving [====================--------------------] 50.00% | 37.80 MB/s
Saved! Hash: 83f4cb...
```

Folders are supported, they are automatically zipped before transfer and the zip is cleaned up afterwards.

---

## Architecture

- **CLI:** Go + [Pion WebRTC v4](https://github.com/pion/webrtc)- single static binary, no runtime dependencies
- **Signaling server:** Go + gorilla/websocket, deployed on Railway
- **Transfer:** WebRTC DataChannel with DTLS encryption, back-pressure to keep memory usage constant
- **Distribution:** GitHub Releases + one-line install scripts for Linux, macOS and Windows

---

## Roadmap

- [x] Word generator for human-readable room codes
- [x] WebSocket signaling server
- [x] CLI sender & receiver
- [x] Folder support (auto-zip)
- [x] SHA-256 hash verification
- [x] Deployed signaling server (Railway)
- [x] Cross-platform releases via GitHub Actions
- [x] One-line install scripts
- [ ] Web client
- [ ] Desktop GUI

---

## License

MIT
