package book

import (
	"context"

	"github.com/firebase/genkit/go/ai"
)

type Repository interface {
	Search(query string) ([]Book, error)
	SemanticSearch(ctx context.Context, bookId string, embeddingVector []float32, n int) ([]*ai.Document, error)
	SaveDocuments(ctx context.Context, bookId string, docs []*ai.Document, vectors [][]float32) error
}
