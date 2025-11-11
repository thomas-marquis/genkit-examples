package chat

import (
	"context"
	"genkit-examples/internal/book"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

const (
	bookRetrieverName = "chatBookRetriever"
)

func (f *Flow) bookRetrieverHandler(ctx context.Context, request *ai.RetrieverRequest) (*ai.RetrieverResponse, error) {
	var bookID string
	if b, ok := request.Query.Metadata["book"].(*book.Book); ok {
		bookID = b.ID
	}

	// Create the embedding vector corresponding to the user request
	res, err := genkit.Embed(ctx, f.g,
		ai.WithDocs(request.Query),
		ai.WithEmbedderName(f.embeddingModelName),
	)
	if err != nil {
		return nil, err
	}

	// Retrieve documents accordingly
	docs, err := f.bookRepository.SemanticSearch(ctx, bookID, res.Embeddings[0].Embedding, 10)
	if err != nil {
		return nil, err
	}

	return &ai.RetrieverResponse{
		Documents: docs, // a utility function to convert chunks to documents
	}, nil
}
