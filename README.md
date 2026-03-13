    <div align="center">

# 🟢 GoChat CLI

### **Peer-to-Peer LAN Chat & File Transfer**

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![Platform](https://img.shields.io/badge/Platform-Windows%20|%20Linux-00ADD8?style=for-the-badge)](#)
[![License](https://img.shields.io/badge/Zero-Dependencies-2EA043?style=for-the-badge)](#)

> **Chat and share files on your local network — no internet, no server, no accounts.**

</div>

---

## 🟢 What Is This?

GoChat is a **command-line P2P chat tool** built in Go. It lets users on the same LAN **discover each other automatically** via UDP broadcast and exchange **messages & files** over TCP — with zero external dependencies.

### Core Concepts Demonstrated

```
 TCP ─────── Reliable message & file delivery
 UDP ─────── Broadcast-based peer discovery
 Goroutines ─ Concurrent server, sender & discovery
 Mutexes ──── Thread-safe peer table
 Sockets ──── Low-level network programming
```

---

## 🟢 Architecture

```
╔══════════════════════════════════════════════╗
║               GoChat Peer                    ║
║                                              ║
║   ┌──────────┐  ┌───────────┐  ┌──────────┐ ║
║   │TCP Server│  │ UDP Bcast │  │UDP Listen│ ║
║   │ :8080    │  │  :9999    │  │  :9999   │ ║
║   └──────────┘  └───────────┘  └──────────┘ ║
║                                              ║
║   ┌──────────────────────────────────────┐   ║
║   │  Peer Table: user → {IP, Port, TTL} │   ║
║   └──────────────────────────────────────┘   ║
╚══════════════════════════════════════════════╝
```

**Flow:** Start → Broadcast username via UDP → Others discover you → TCP connection for messages/files

---

## 🟢 Quick Start

```bash
# Build
go build -o gochat.exe .

# Add to PATH (one-time, PowerShell)
[Environment]::SetEnvironmentVariable("Path", $env:PATH + ";C:\gochatcli", "User")
```

---

## 🟢 Commands

| Command | What It Does |
|:--------|:-------------|
| `gochat -listen -u basil` | 📡 Listen for messages & announce presence |
| `gochat -chat -u basil -ip 172.17.3.175` | 💬 Interactive live chat |
| `gochat -u alice -m "Hello!"` | ✉️ Send a message (auto-discover) |
| `gochat -u alice -ip 10.0.0.5 -m "Hi"` | 📌 Send message (direct IP) |
| `gochat -u alice -t ./file.pdf` | 📁 Send a file |
| `gochat -users` | 👥 List online users |

### Flags

| Flag | Description | Default |
|:-----|:------------|:--------|
| `-listen` | Listener mode | `false` |
| `-chat` | Interactive chat | `false` |
| `-u` | Username | *required* |
| `-m` | Message text | — |
| `-t` | File path | — |
| `-ip` | Direct IP (skip discovery) | — |
| `-port` | TCP port | `8080` |
| `-users` | Show online users | `false` |

---

## 🟢 File Structure

| File | Role |
|:-----|:-----|
| `main.go` | Entry point & CLI flag parsing |
| `chat.go` | Interactive chat mode |
| `server.go` | TCP server — receives messages & files |
| `client.go` | TCP client — sends messages & files |
| `discovery.go` | UDP broadcast & peer discovery |
| `peer.go` | Thread-safe peer table |
| `sockopt_*.go` | Platform-specific socket options |

---

## 🟢 Example

**Machine A** — start listening:
```
> gochat -listen -u basil
[gochat] starting as 'basil' on TCP port 8080
[discovery] found peer: alice @ 172.17.3.180:8080
[message from alice]: Hey Basil!
```

**Machine B** — send a message:
```
> gochat -u basil -m "Hey Basil!"
[discovery] found basil @ 172.17.3.175:8080
[sent] message to 172.17.3.175:8080
```

---

<div align="center">

**Built with 🟢 Go** • **Zero Dependencies** • **Pure Standard Library**

</div>
