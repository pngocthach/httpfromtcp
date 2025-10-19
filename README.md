# HTTP from TCP: Go HTTP Server from Scratch 🚀

## Description 📖

A custom HTTP/1.1 server built in Go using only raw TCP connections. This project focuses on implementing the core HTTP parsing logic (request line, headers, body) and response generation without relying on Go's standard `net/http` package, demonstrating a low-level understanding of the protocol. It includes features like concurrent connection handling, a streaming parser with a state machine, basic response generation, custom handler support, and graceful shutdown.

---

## Tech Stack 🛠️

- **Go** (>= 1.18)
- Go Standard Libraries (`net`, `io`, `bytes`, `sync/atomic`, etc.)
- Testify (`github.com/stretchr/testify`) for testing.

---

## Installation & Usage ⚙️▶️

1.  **Clone the repository:**

```bash
git clone https://github.com/pngocthach/httpfromtcp
cd httpfromtcp
```

2.  **Install dependencies:**

```bash
go mod tidy
```

3.  **Build and run the server:** (Listens on port `3000`)

```bash
go build -o bin/httpserver ./cmd/tcplistener
./bin/httpserver
```

4.  **Test the server:**

```bash
curl -v http://localhost:3000/some/path
```

---

## Testing ✅

Run all unit tests from the project root:

```bash
go test ./...
```

---

## What I Learned 🧠

- Deep dive into **HTTP/1.1 protocol internals** (RFC 9110, 9112).
- Handling **TCP streams** and chunked data.
- Building **parsers** and **state machines** in Go.
- Effective **buffer management** (resizing, compaction).
- **Concurrency** with Goroutines.
- **TDD** using Testify.

---

## Acknowledgements 🙏

Inspired and guided by the ["Learn HTTP"](https://boot.dev/courses/learn-http-protocol-golang) course on **Boot.dev** and content from **ThePrimeagen**.
