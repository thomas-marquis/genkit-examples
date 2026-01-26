# Genkit Examples

A Go-based AI application demonstrating Retrieval Augmented Generation (RAG) with a vector database and a simple chat API. It uses a YAML configuration file, a PostgreSQL database (with pgvector), and exposes an HTTP endpoint to ask questions.

This project is built with Firebase Genkit for defining AI flows, a custom retriever backed by pgvector in Postgres, and an HTTP server exposing a single POST /ask endpoint.

- LLM/Embeddings: Mistral (via github.com/thomas-marquis/genkit-mistral)
- Vector DB: PostgreSQL + pgvector
- Tests: go test with Testcontainers (spawns a temporary pgvector instance)

## Prerequisites

- Go 1.25+ (see go.mod)
- Docker (required for Postgres locally and also for running tests via Testcontainers)
- curl (for trying the API)

Optional but handy:
- Genkit CLI (for the dev UI): `curl -sL cli.genkit.dev | bash` or see Genkit docs

## Configuration

The app reads settings.yaml from the repository root via Viper.
Use settings-samples.yaml as a template:

Notes:
- Keep secrets out of version control. Do not commit your personal settings.yaml.

## Start a local Postgres (with pgvector)

A docker-compose file is provided.

```
docker compose up -d
```

It starts a database at:
- Host: localhost
- Port: 5432
- User: postgres
- Password: password
- Database: mydb

Ensure settings.yaml matches these values (or adjust docker-compose.yaml accordingly).

## Running the application

```
go run .
```

The HTTP server listens on 127.0.0.1:3400 and exposes:
- POST /ask — calls the Genkit "chat-flow" which retrieves relevant chunks and generates an answer.

Optional: Start the Genkit Dev UI alongside the app (requires the Genkit CLI installed). A Make target is provided:

```
make dev-ui
```

This will run `genkit start -- go run .` and open the Genkit UI for inspecting flows.

## Seeding/Indexing your knowledge base

The chat flow relies on a pgvector-backed retriever. You need to populate the vector store with content first.

The Agent includes an Index method that:
- Parses EPUB content to Markdown
- Splits it into chunks
- Embeds each chunk (with Mistral embedder)
- Inserts into Postgres with pgvector

You can write a small Go script to index one or more EPUB files:

```go
package main

import (
  "context"
  "genkit-examples/internal/agent"
  "genkit-examples/internal/book"
  "genkit-examples/internal/vectorstore"
  "github.com/firebase/genkit/go/genkit"
)

func main() {
  ctx := context.Background()
  g := genkit.Init(ctx) // for indexing we only need embedding capability; ensure your Mistral plugin is configured like in main.go if running standalone

  // Create your vector store (adjust connection string to your settings)
  conn := "postgres://postgres:password@localhost:5432/mydb"
  vs, _ := vectorstore.New(conn)

  a := agent.New(ctx, g, vs)
  books := []book.Book{{
    ID:       "my-epub-1",
    Title:    "My EPUB",
    FilePath: "/path/to/book.epub",
  }}
  _ = a.Index(ctx, books)
}
```

Alternatively, see internal/agent/agent_test.go for a fully self-contained example of preparing the store and exercising the flow.

## API usage

```
curl -sS \
  -X POST http://127.0.0.1:3400/ask \
  -H "Content-Type: application/json" \
  -d '{"question":"What is Go programming language?"}'
```

Example successful response:

```
{"answer":"Go is a programming language developed by Google."}
```

## Running tests

Tests require Docker because they use Testcontainers to spin up a temporary pgvector Postgres.

Run all tests:

```
go test ./...
```

The tests mock the LLM and embedding models, so you don’t need real API keys to run them.

## Troubleshooting

- replace directive in go.mod: If you see module resolution errors for github.com/thomas-marquis/genkit-mistral, remove or adjust the replace line at the bottom of go.mod.
- Database mismatches: Verify settings.yaml db.name matches docker-compose (default: mydb) and the app can reach 127.0.0.1:5432.
- No answers / errors from /ask: Ensure you indexed content into the vector store before asking questions.
