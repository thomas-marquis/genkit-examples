package agent

import (
	"context"
	"fmt"
	"genkit-examples/internal/book"
	"log"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

func (a *Agent) Index(ctx context.Context, books []book.Book) error {
	for i, b := range books {
		log.Printf("Adding book #%d...", i)
		chunks, err := b.Parse()
		if err != nil {
			return err
		}

		if err := batchAndApply(chunks, 10, func(startIdx int, batch []*ai.Document) error {
			log.Printf("Batch %d/%d...", startIdx, len(chunks))
			embeddings, err := embedChunks(ctx, a.g, batch)
			if err != nil {
				return fmt.Errorf("failed to embed documents for batch at start index %d: %w", startIdx, err)
			}

			log.Println("Embedding done, inserting chunks...")
			if err := a.vecStore.InsertDocuments(ctx, b, batch, embeddings); err != nil {
				return fmt.Errorf("failed to insert chunks for batch at start index %d: %w", startIdx, err)
			}
			return nil
		}); err != nil {
			return err
		}

		log.Println("book added")
	}

	return nil
}

func embedChunks(ctx context.Context, g *genkit.Genkit, chunks []*ai.Document) ([]*ai.Embedding, error) {
	embeddings := make([]*ai.Embedding, 0, len(chunks))
	for _, chunk := range chunks {
		res, err := genkit.Embed(ctx, g,
			ai.WithEmbedderName("mistral/mistral-embed"),
			ai.WithDocs(chunk),
		)
		if err != nil {
			return nil, err
		}
		embeddings = append(embeddings, res.Embeddings[0])
	}

	return embeddings, nil
}

// batchAndApply splits the provided slice into batches of at most batchSize and the apply the provided function to each batch.
// The last batch may contain fewer elements. For non-positive batchSize, it defaults to 1.
func batchAndApply[T any](data []T, batchSize int, fn func(int, []T) error) error {
	if batchSize <= 0 {
		batchSize = 1
	}

	if fn == nil {
		fn = func(startIdx int, batch []T) error {
			return nil
		}
	}

	n := len(data)
	if n == 0 {
		return nil
	}

	for i := 0; i < n; i += batchSize {
		end := i + batchSize
		if end > n {
			end = n
		}
		if err := fn(i, data[i:end]); err != nil {
			return err
		}
	}
	return nil
}
