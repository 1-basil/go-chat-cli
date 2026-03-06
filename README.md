# GoChat CLI — LAN Chat & File Transfer Tool

## 📋 Project Overview (For Teachers)

**GoChat CLI** is a command-line peer-to-peer (P2P) chat application written in **Go (Golang)**. It allows users on the **same Local Area Network (LAN)** to discover each other automatically and exchange **text messages** and **files** — without needing any central server.

### What Problem Does It Solve?

In traditional messaging, you need an internet connection and a centralized server (like WhatsApp or Telegram). GoChat eliminates both requirements:

- **No internet needed** — works entirely over the local network (e.g., a college lab, home Wi-Fi)
- **No central server** — every machine acts as both a client and a server (true P2P)
- **No setup or accounts** — just pick a username and start chatting

### What Are We Exactly Doing?

We are building a **networking application** that demonstrates core concepts of:

| Concept | How GoChat Uses It |
|---|---|
| **TCP (Transmission Control Protocol)** | Reliable delivery of messages and files between peers |
| **UDP (User Datagram Protocol)** | Broadcasting presence to discover other users on the LAN |
| **Socket Programming** | Opening, listening on, and writing to network sockets in Go |
| **Concurrency (Goroutines)** | Running the server, discovery broadcaster, and discovery listener simultaneously |
| **Thread Safety (Mutexes)** | Safely reading/writing the peer table from multiple goroutines |
| **CLI Design (Flags)** | Using command-line flags for a clean, Unix-style user interface |
| **File I/O & Streaming** | Reading files from disk and streaming them over TCP connections |

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────┐
│                   GoChat Peer                    │
│                                                  │
│  ┌────────────┐  ┌─────────────┐  ┌───────────┐ │
│  │ TCP Server │  │ UDP         │  │ UDP       │ │
│  │ (Receive   │  │ Broadcaster │  │ Listener  │ │
│  │  msgs/     │  │ (Announce   │  │ (Discover │ │
│  │  files)    │  │  presence)  │  │  peers)   │ │
│  │ Port 8080  │  │ Port 9999   │  │ Port 9999 │ │
│  └────────────┘  └─────────────┘  └───────────┘ │
│                                                  │
│  ┌──────────────────────────────────────────────┐│
│  │           Peer Table (In-Memory)             ││
│  │  username → { IP, Port, LastSeen }           ││
│  └──────────────────────────────────────────────┘│
└──────────────────────────────────────────────────┘
```

### How It Works (Step by Step)

1. **User starts GoChat** with a username (e.g., `basil`)
2. **UDP Broadcaster** sends `DISCOVER:basil:8080` packets to the entire LAN every 5 seconds
3. **UDP Listener** on other machines picks up these broadcasts and adds `basil` to their **Peer Table**
4. When someone wants to message `basil`, they look up his IP from the Peer Table
5. A **TCP connection** is opened to `basil`'s machine, and the message/file is sent reliably

---

## 📁 Project File Structure

| File | Purpose |
|---|---|
| `main.go` | Entry point — parses CLI flags and dispatches to the correct mode |
| `chat.go` | Interactive chat mode — read from stdin, send messages in real-time |
| `server.go` | TCP server — receives incoming messages and files from peers |
| `client.go` | TCP client — sends messages (`SendMessage`) and files (`SendFile`) to peers |
| `discovery.go` | UDP broadcast & listen logic for automatic peer discovery |
| `peer.go` | Peer data structure and thread-safe Peer Table |
| `sockopt_windows.go` | Windows-specific socket options (SO_REUSEADDR) |
| `sockopt_other.go` | Linux/macOS socket options |
| `go.mod` | Go module definition |

---

## 🚀 How to Use

### Prerequisites

- **Go** installed (version 1.25+)
- Two or more computers on the **same LAN** (or same machine for testing)

### Build

```bash
go build -o gochat.exe .
```

### Commands

#### 1. Listen Mode — Wait for Messages & Announce Presence

```bash
.\gochat.exe -listen -u <your_username>
```

**Example:**
```bash
.\gochat.exe -listen -u basil
```

This starts a TCP server on port 8080 and announces your presence via UDP. You'll see incoming messages and files here.

---

#### 2. Send a Message to a User (Auto-Discovery)

```bash
.\gochat.exe -u <target_username> -m "your message here"
```

**Example:**
```bash
.\gochat.exe -u alice -m "Hey Alice, how are you?"
```

GoChat will search the network for `alice` for up to 5 seconds. If found, the message is delivered instantly.

---

#### 3. Send a Message by Direct IP (Skip Discovery)

```bash
.\gochat.exe -u <target_username> -ip <target_ip> -m "message"
```

**Example:**
```bash
.\gochat.exe -u alice -ip 172.17.3.175 -m "Hello from Basil!"
```

---

#### 4. Send a File

```bash
.\gochat.exe -u <target_username> -t <file_path>
```

**Example:**
```bash
.\gochat.exe -u alice -t ./report.pdf
```

You can also combine with `-ip` to skip discovery.

---

#### 5. Interactive Chat Mode

```bash
.\gochat.exe -chat -u <your_username> -ip <peer_ip>
```

**Example:**
```bash
.\gochat.exe -chat -u basil -ip 172.17.3.175
```

Opens a live chat session. Type messages and press Enter to send. Press `Ctrl+C` to exit.

---

#### 6. List Online Users

```bash
.\gochat.exe -users
```

Scans the network for 3 seconds and displays all discovered users with their IP addresses.

---

### All CLI Flags

| Flag | Description | Default |
|---|---|---|
| `-listen` | Start in listener mode (receive messages) | `false` |
| `-chat` | Start interactive chat mode | `false` |
| `-u` | Username (yours or target's, depending on mode) | *(required)* |
| `-m` | Text message to send | — |
| `-t` | File path to send | — |
| `-ip` | Direct IP address (skips auto-discovery) | — |
| `-port` | TCP port for the server | `8080` |
| `-users` | Discover and show online users | `false` |

---

## 🔧 Technologies & Concepts Used

- **Language:** Go (Golang)
- **Networking:** TCP for reliable data transfer, UDP for broadcast discovery
- **Concurrency:** Goroutines and sync.RWMutex for thread-safe operations
- **Platform Support:** Cross-compiled for Windows and Linux (build tag files for socket options)
- **No external dependencies** — uses only Go's standard library (`net`, `os`, `fmt`, `sync`, `bufio`, etc.)

---

## 📝 Example Session

**Machine A (Basil) — Start listening:**
```
> .\gochat.exe -listen -u basil
[gochat] starting as 'basil' on TCP port 8080
[server] listening on TCP port 8080
[discovery] broadcasting to 3 address(es)
[discovery] found peer: alice @ 172.17.3.180:8080

[message from alice]: Hey Basil!
[file from alice]: receiving 'notes.txt' (1024 bytes)...
[file from alice]: saved 'notes.txt' (1024 bytes)
```

**Machine B (Alice) — Send a message and file:**
```
> .\gochat.exe -u basil -m "Hey Basil!"
[discovery] looking for 'basil' on the network...
[discovery] found basil @ 172.17.3.175:8080
[sent] message to 172.17.3.175:8080

> .\gochat.exe -u basil -t notes.txt
[discovery] looking for 'basil' on the network...
[discovery] found basil @ 172.17.3.175:8080
[sent] file 'notes.txt' (1024 bytes) to 172.17.3.175:8080
```
