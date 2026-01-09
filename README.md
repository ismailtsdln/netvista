# 👁️ NetVista v2.1.0-pro

<p align="center">
  <img src="web/assets/logo.png" alt="NetVista Logo" width="200">
</p>

**NetVista v2.1.0-pro** is a professional-grade Visual Reconnaissance tool refactored with **Hexagonal Architecture (Ports & Adapters)** for enterprise-scale reliability. It provides deep visibility into your target's attack surface through smart rendering, technology fingerprinting, and visual intelligence.

## ✨ Advanced Features (v2.0)

- 🏗️ **Hexagonal Architecture**: Decoupled core logic from infrastructure for extreme modularity and testability.
- 📸 **Smart Rendering**: Adaptive waiting (NetworkIdle) and automatic detection/bypass of cookie consent overlays.
- 🧠 **Framework Intelligence**: Automatic detection of modern SPA frameworks (React, Vue, etc.).
- 🔄 **Incremental Scanning**: Intelligently skips previously processed targets to save time and resources.
- ⚡ **Resource-Aware Scaling**: Thread-safe browser context pooling and responsive concurrency management.
- 🔍 **Visual Deduplication**: Uses pHash to group visually identical hosts, reducing report noise.
- 🛡️ **Advanced Resilience**: Exponential backoff and automated retries for transient network failures.
- 📊 **Enterprise Reporting**: Multi-format exports (JSON, CSV, Markdown, Text, ZIP) and Interactive Dashboard 2.0.

## 🚀 Quick Start

### Installation

### From Source

```bash
git clone https://github.com/ismailtsdln/netvista
cd netvista
go build -o netvista cmd/netvista/main.go
```

### Basic Scan

```bash
echo "example.com" | ./netvista scan -o my_recon
```

### Incremental Scan (Resuming)

```bash
echo "example.com" | ./netvista scan -o my_recon  # Skips if already in results.json
```

## 🏗️ Architecture (Hexagonal)

NetVista v2.0 follows the Clean Architecture / Ports & Adapters pattern:

- **Core/Domain**: Pure business logic and models.
- **Core/Ports**: Interface definitions for infra dependencies (Prober, Renderer, Reporter).
- **Core/Services**: Orchestration layer (The `ScannerService`).
- **Infra/Adapters**: Implementations using external tools (Playwright, Net/HTTP, Go-Report).

## 📊 Interaction Dashboard
Launch the dashboard to explore your scan results visually:
```bash
./netvista serve -d my_recon
```

---
Created with ❤️ by **İsmail Taşdelen**
