package vectorstore

import (
	"context"
	"errors"
	"genkit-examples/internal/book"

	"github.com/firebase/genkit/go/ai"
	"github.com/pgvector/pgvector-go"
	"github.com/thomas-marquis/genkit-mistral/mistral"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VectorStore struct {
	db *gorm.DB
}

func New(connStr string) (*VectorStore, error) {
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector;").Error; err != nil {
		return nil, err
	}

	x := &Chunk{}

	if err := db.AutoMigrate(x); err != nil {
		return nil, err
	}

	return &VectorStore{db: db}, nil
}

func (s *VectorStore) InsertDocuments(ctx context.Context, b book.Book, chunks []*ai.Document, embeddings []*ai.Embedding) error {
	if len(chunks) != len(embeddings) {
		return errors.New("chunks and embeddings must have the same length")
	}

	chunksDTO := make([]*Chunk, 0, len(chunks))
	for i, chunk := range chunks {
		emb := embeddings[i]
		c := Chunk{
			Embedding: pgvector.NewVector(emb.Embedding),
			Content:   mistral.StringFromParts(chunk.Content),
			BookID:    b.ID,
		}
		chunksDTO = append(chunksDTO, &c)
	}

	return s.db.
		WithContext(ctx).
		Clauses(clause.OnConflict{ // Ensure a given chunk is not added twice: duplication may decrease the RAG's relevance
			Columns:   []clause.Column{{Name: "book_id"}, {Name: "content"}},
			DoUpdates: clause.AssignmentColumns([]string{"embedding"}),
		}).
		Create(deduplicateChunks(chunksDTO)).Error // We need to also deduplicate at the batch level
}

func deduplicateChunks(chunks []*Chunk) []*Chunk {
	deduplicated := make([]*Chunk, 0, len(chunks))
	seen := make(map[string]bool)
	for _, chunk := range chunks {
		if !seen[chunk.Content] {
			seen[chunk.Content] = true
			deduplicated = append(deduplicated, chunk)
		}
	}
	return deduplicated
}
