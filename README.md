# 👁️ NetVista

<p align="center">
  <img src="web/assets/logo.png" alt="NetVista Logo" width="200">
</p>

**NetVista** is a modern, high-performance Visual Reconnaissance tool designed for security researchers and bug bounty hunters. It combines network probing, technology fingerprinting, and visual intelligence to provide a comprehensive view of your target's attack surface.

## ✨ Features

- 🚀 **Asynchronous Discovery**: Ultra-fast HTTP probing and port scanning.
- 📸 **Visual Intelligence**: Automated screenshot capturing using Playwright.
- 🧠 **Perceptual Hashing (pHash)**: Automatically groups visually similar hosts to minimize redundancy.
- 🛡️ **Subdomain Takeover Detection**: Built-in signature-based detection for vulnerable cloud services.
- 🔍 **Advanced Fingerprinting**: Identifies technologies via headers, titles, and body patterns.
- 📊 **Interactive Dashboard**: Modern, searchable, and filterable HTML report with Dark/Light mode support.
- 🗜️ **Report Distribution**: Automatic ZIP export for easy report sharing.
- 🌐 **Proxy & Header Support**: HTTP/SOCKS5 proxy support and custom header injection.
- ⚙️ **YAML Configuration**: Persistent settings via `netvista.yaml`.

## 🚀 Quick Start

### Installation

#### Install via Go (Recommended)
```bash
go install github.com/ismailtsdln/netvista/cmd/netvista@latest
```

#### From Source
```bash
git clone https://github.com/ismailtsdln/netvista
cd netvista
go build -o netvista cmd/netvista/main.go
```


### Usage

#### Basic Scan
```bash
echo "example.com" | ./netvista scan
```

#### Nmap Integration
```bash
./netvista scan --nmap targets.xml -o my_report
```

#### Serve Dashboard
```bash
./netvista serve -d my_report
```

## 🛠️ Configuration (`netvista.yaml`)

```yaml
ports: "80,443,8000,8080,8443"
concurrency: 20
output: "reports"
timeout: "10s"
proxy: ""
headers: "User-Agent: NetVista-Scanner"
```

## 🏗️ Architecture

NetVista is built with modularity in mind:
- **Engine**: Handles the worker pool and target distribution.
- **Prober**: Performs HTTP/S probing and metadata extraction.
- **Capturer**: Manages a headless browser pool for high-fidelity screenshots.
- **Plugins**: Extensible system for fingerprinting and vulnerability detection.
- **Reporting**: Generates static JSON results and dynamic HTML dashboards.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
Created with ❤️ by **İsmail Taşdelen**
