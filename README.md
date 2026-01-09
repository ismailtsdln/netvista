# 🚀 NetVista — Network & Web Host Visual Recon Tool

NetVista is a modern, scalable, and modular visual reconnaissance tool designed to help security professionals and researchers quickly visualize large-scale networks and web host landscapes. Built with Go for performance and Playwright for robust rendering, NetVista bridges the gap between raw scan data and actionable visual intelligence.

---

## ✨ Key Features

-   **🔍 Flexible Input**: Accepts piped input (DNS, IP, host lists) or Nmap/Masscan XML files.
-   **⚡ High-Performance Scanning**: Asynchronous port scanning and HTTP probing with customizable concurrency and rate limiting.
-   **📸 Smart Rendering**: Integrated Playwright screenshot engine handles modern, JS-heavy single-page applications (SPAs) with ease.
-   **🧩 Modular Plugin Architecture**: Easily extendable for technology fingerprinting, domain takeover detection, and metadata extraction.
-   **📊 Visual Clustering**: Group similar hosts using perceptual hashing (pHash) to reduce noise in large assessments.
-   **📈 Modern Reporting**: Generates a clean, filterable HTML dashboard with search, pagination, and export options (JSON, CSV).
-   **🛠 Resilience**: Built-in retry logic, backoff strategies, and graceful shutdown handling.

---

## 🚀 Installation

### Prerequisites

-   [Go](https://golang.org/doc/install) (1.21+)
-   [Playwright for Go](https://github.com/playwright-community/playwright-go) (The tool will attempt to install browsers automatically)

### Build from Source

```bash
git clone https://github.com/ismailtsdln/netvista.git
cd netvista
go build -o netvista cmd/netvista/main.go
```

---

## 🛠 Usage & Examples

### Basic Scan (Piped Input)

```bash
cat targets.txt | ./netvista scan
```

### Import from Nmap XML

```bash
./netvista --nmap scan.xml --output reports/
```

### Advanced Usage with Concurrency Control

```bash
./netvista scan --concurrency 50 --timeout 10s --ports 80,443,8080-8090
```

---

## 🧩 Plugins

NetVista supports a variety of submodules to enrich scan data:

-   **Tech Fingerprinting**: Identify CMS, frameworks, and web servers (Wappalyzer integration).
-   **Domain Takeover**: Check for dangling DNS records pointing to unclaimed third-party services.
-   **Metadata Extraction**: Capture HTTP headers, page titles, and body snippets.

---

## 📝 Configuration

Configuration can be managed via command-line flags or a `config.yaml` file (coming soon).

| Flag | Description | Default |
| :--- | :--- | :--- |
| `-p, --ports` | Ports to scan (e.g., 80,443,8000-9000) | `80,443` |
| `-c, --concurrency` | Number of concurrent workers | `20` |
| `-o, --output` | Output directory for reports | `./out` |
| `-t, --timeout` | Timeout per host | `5s` |

---

## 📊 Reports

The generated HTML dashboard includes:
-   **Screenshot Gallery**: Grid view of all captured screenshots.
-   **Host details**: Status codes, page titles, and technology stack.
-   **Interactive Filters**: Filter by response code, technology, or port.

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the Branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request

---

## ⚖️ License

Distributed under the MIT License. See `LICENSE` for more information.

---

## 👨‍💻 Author

**İsmail Taşdelen**
-   GitHub: [@ismailtsdln](https://github.com/ismailtsdln)
-   Project Link: [NetVista](https://github.com/ismailtsdln/netvista)
