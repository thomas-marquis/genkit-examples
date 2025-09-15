package agent_test

import (
	"context"
	"fmt"
	"genkit-examples/internal/agent"
	"genkit-examples/internal/book"
	"genkit-examples/internal/vectorstore"
	"testing"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupTestContainer(t *testing.T, ctx context.Context) (*postgres.PostgresContainer, string, error) {
	t.Helper()

	dbName := "testdb"
	dbUser := "testuser"
	dbPassword := "testpass"

	postgresContainer, err := postgres.Run(ctx,
		"pgvector/pgvector:pg16",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, "", err
	}

	// Get database connection details
	host, err := postgresContainer.Host(ctx)
	if err != nil {
		return nil, "", err
	}
	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, "", err
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, host, port.Port(), dbName,
	)

	return postgresContainer, connStr, nil
}

func setupVectorStore(t *testing.T, ctx context.Context, vs *vectorstore.VectorStore) {
	t.Helper()

	testChunks := []*ai.Document{
		ai.DocumentFromText("Go is a programming language developed by Google. It's known for its simplicity and efficiency.", nil),
		ai.DocumentFromText("Artificial Intelligence is transforming how we solve complex problems across various industries.", nil),
		ai.DocumentFromText("Machine learning algorithms can learn patterns from data without explicit programming.", nil),
		ai.DocumentFromText("Vector databases are essential for storing and retrieving high-dimensional embeddings efficiently.", nil),
		ai.DocumentFromText("The RAG (Retrieval Augmented Generation) pattern combines retrieval with generation for better AI responses.", nil),
		ai.DocumentFromText("The Mistral model is a powerful AI model that can generate text in a multiturn conversation.", nil),
		ai.DocumentFromText("The Mistral embedder is a powerful AI embedder that can generate embeddings for text in a multiturn conversation.", nil),
	}
	fakeBook := book.Book{ID: "test-book-123", Title: "Test Book"}

	fakeEmbeddings := make([]*ai.Embedding, len(testChunks))
	for i := range testChunks {
		fakeEmbeddings[i] = &ai.Embedding{
			Embedding: make([]float32, 1024),
		}
	}

	if err := vs.InsertDocuments(ctx, fakeBook, testChunks, fakeEmbeddings); err != nil {
		t.Fatal("failed to insert test documents:", err)
	}
}
func setupEmbedding(g *genkit.Genkit) {
	genkit.DefineEmbedder(g, "mistral/mistral-embed", &ai.EmbedderOptions{},
		func(ctx context.Context, request *ai.EmbedRequest) (*ai.EmbedResponse, error) {
			embeddings := make([]*ai.Embedding, len(request.Input))
			for i := range request.Input {
				// Create fake 1024-dimensional embeddings
				fakeEmbedding := make([]float32, 1024)
				embeddings[i] = &ai.Embedding{Embedding: fakeEmbedding}
			}
			return &ai.EmbedResponse{Embeddings: embeddings}, nil
		})
}

func Test_ChatFlow(t *testing.T) {
	ctx := context.Background()

	postgresContainer, connStr, err := setupTestContainer(t, ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			t.Fatal("failed to terminate container:", err)
		}
	}()
	if err != nil {
		t.Fatal("failed to setup test container:", err)
	}

	vs, err := vectorstore.New(connStr)
	if err != nil {
		t.Fatal("failed to create vector store:", err)
	}

	setupVectorStore(t, ctx, vs)

	// Mock embedder with fake embeddings (common for all tests)
	setupEmbedding(g)

	t.Run("should respond with text", func(t *testing.T) {
		// Given
		g := genkit.Init(ctx) // Create an empty Genkit instance, without any plugin to be sure we don't have side effects
		a := agent.New(ctx, g, vs)

		// Mock model with a fake response. This fake model will be called by Genkit under the hood.
		genkit.DefineModel(g, "mistral/mistral-small-latest",
			&ai.ModelOptions{Supports: &ai.ModelSupports{Multiturn: true}},
			func(ctx context.Context, request *ai.ModelRequest, s core.StreamCallback[*ai.ModelResponseChunk]) (*ai.ModelResponse, error) {
				return &ai.ModelResponse{
					Message: ai.NewTextMessage(ai.RoleModel, "Go is a programming language developed by Google."),
				}, nil
			})

		// When
		response, err := a.Ask("What is Go programming language?")

		// Then
		assert.NoError(t, err, "unexpected error from Ask")
		assert.Equal(t, "Go is a programming language developed by Google.", response)
	})

	t.Run("should retrieve and send 5 documents to the model", func(t *testing.T) {
		// Given
		g := genkit.Init(ctx)
		a := agent.New(ctx, g, vs)

		genkit.DefineModel(g, "mistral/mistral-small-latest",
			&ai.ModelOptions{Supports: &ai.ModelSupports{Multiturn: true}},
			func(ctx context.Context, request *ai.ModelRequest, s core.StreamCallback[*ai.ModelResponseChunk]) (*ai.ModelResponse, error) {
				assert.Len(t, request.Docs, 5, "expected 5 documents to be sent to the model")

				return &ai.ModelResponse{
					Message: ai.NewTextMessage(ai.RoleModel, "Go is a programming language developed by Google."),
				}, nil
			})

		// When
		response, err := a.Ask("What is Go programming language?")

		// Then
		assert.NoError(t, err, "unexpected error from Ask")
		assert.Equal(t, "Go is a programming language developed by Google.", response)
	})
}
