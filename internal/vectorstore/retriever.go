package vectorstore

import (
	"context"
	"genkit-examples/internal/book"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm/clause"
)

func (s *VectorStore) MakeRetrieverHandler(g *genkit.Genkit) func(context.Context, *ai.RetrieverRequest) (*ai.RetrieverResponse, error) {
	return func(ctx context.Context, request *ai.RetrieverRequest) (*ai.RetrieverResponse, error) {
		var bookID string
		if book, ok := request.Query.Metadata["book"].(*book.Book); ok {
			bookID = book.ID
		}

		// Create the embedding vector corresponding to the user request
		res, err := genkit.Embed(ctx, g,
			ai.WithDocs(request.Query),
			ai.WithEmbedderName("mistral/mistral-embed"),
		)
		if err != nil {
			return nil, err
		}

		// Retrieve documents accordingly
		var retrievedChunks []*Chunk
		if err := s.db.
			WithContext(ctx).
			Clauses(clause.OrderBy{
				Expression: clause.Expr{
					SQL:  "embedding <=> ?", // PG vector's cosine similarity syntax
					Vars: []any{pgvector.NewVector(res.Embeddings[0].Embedding)},
				},
			}).
			Where(Chunk{BookID: bookID}).
			Limit(5).
			Find(&retrievedChunks).Error; err != nil {
			return nil, err
		}

		return &ai.RetrieverResponse{
			Documents: toDocuments(retrievedChunks), // a utility function to convert chunks to documents
		}, nil
	}
}
