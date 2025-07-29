# Slayer ⚔️  
> Network Control & ARP Spoofing Tool – Fast, Modular & Written in Go

![License](https://img.shields.io/badge/license-MIT-blue)
![Go Version](https://img.shields.io/badge/go-1.21+-blue)
![Platform](https://img.shields.io/badge/platform-linux-lightgrey)

---

**Slayer** is a powerful command-line tool for local network testing, device rate-limiting, and ARP spoofing. Inspired by tools like `evillimiter` — but built from the ground up in Go — Slayer offers a modular, terminal-first approach with a focus on performance, clarity, and full control.

---

## ⚡ Features

- 🔍 **Network Scanning** via ARP
- 🎯 **Per-host Upload/Download Limiting** using `iptables` + `tc`
- 🕵️ **ARP Spoofing** (man-in-the-middle) with live control
- 📟 **Interactive Shell** with command history and navigation
- 🛠️ Root-level system requirement checks
- 🧠 Lightweight, dependency-minimal design
- 🧼 Graceful shutdown and cleanup

---

## 📦 Installation

### ⚙️ Prerequisites
- **Linux**
- **Go 1.21+**
- Root privileges (`sudo`)
- Required binaries in `$PATH`: `iptables`, `tc`, `ip`

### 🛠 Build from source

```bash
git clone https://github.com/prabalesh/slayer.git
cd slayer
go mod tidy
go build -o slayer ./cmd/slayer
sudo ./slayer
