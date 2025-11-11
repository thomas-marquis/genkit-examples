package infrastructure

import (
	"context"
	"errors"
	"genkit-examples/internal/infrastructure/orm"

	"github.com/firebase/genkit/go/ai"
	"github.com/pgvector/pgvector-go"
	"github.com/thomas-marquis/genkit-mistral/mistral"
	"gorm.io/gorm/clause"
)

func (c *bookRepositoryImpl) SemanticSearch(ctx context.Context, bookId string, embeddingVector []float32, n int) ([]*ai.Document, error) {
	var retrievedChunks []*orm.BookPart
	if err := c.db.
		WithContext(ctx).
		Clauses(clause.OrderBy{
			Expression: clause.Expr{
				SQL:  "embedding <=> ?", // PG vector's cosine similarity syntax
				Vars: []any{pgvector.NewVector(embeddingVector)},
			},
		}).
		Where(orm.BookPart{BookID: bookId}).
		Limit(n).
		Find(&retrievedChunks).Error; err != nil {
		return nil, err
	}

	return orm.DocumentsFromBookParts(retrievedChunks), nil
}

func (c *bookRepositoryImpl) SaveDocuments(ctx context.Context, bookId string, docs []*ai.Document, vectors [][]float32) error {
	if len(docs) != len(vectors) {
		return errors.New("documents and embeddings must have the same length")
	}

	chunksDTO := make([]*orm.BookPart, 0, len(docs))
	for i, doc := range docs {
		emb := vectors[i]
		c := orm.BookPart{
			Embedding: pgvector.NewVector(emb),
			Content:   mistral.StringFromParts(doc.Content),
			BookID:    bookId,
		}
		chunksDTO = append(chunksDTO, &c)
	}

	return c.db.
		WithContext(ctx).
		Clauses(clause.OnConflict{ // Ensure a given doc is not added twice: duplication may decrease the RAG's relevance
			Columns:   []clause.Column{{Name: "book_id"}, {Name: "content"}},
			DoUpdates: clause.AssignmentColumns([]string{"embedding"}),
		}).
		Create(deduplicateChunks(chunksDTO)).Error // We need to also deduplicate at the batch level
}

func deduplicateChunks(chunks []*orm.BookPart) []*orm.BookPart {
	deduplicated := make([]*orm.BookPart, 0, len(chunks))
	seen := make(map[string]bool)
	for _, chunk := range chunks {
		if !seen[chunk.Content] {
			seen[chunk.Content] = true
			deduplicated = append(deduplicated, chunk)
		}
	}
	return deduplicated
}

func (c *bookRepositoryImpl) migrate() error {
	if err := c.db.Exec("CREATE EXTENSION IF NOT EXISTS vector;").Error; err != nil {
		return err
	}

	migrator := c.db.Migrator()

	if !migrator.HasTable(&orm.Book{}) {
		if err := migrator.CreateTable(&orm.Book{}); err != nil {
			return err
		}
	}

	if !migrator.HasTable(&orm.BookPart{}) {
		if err := migrator.CreateTable(&orm.BookPart{}); err != nil {
			return err
		}
	} else {
		if !migrator.HasIndex(&orm.BookPart{}, "idx_book_content") {
			if err := migrator.AlterColumn(&orm.BookPart{}, "Content"); err != nil {
				return err
			}
			if err := migrator.AlterColumn(&orm.BookPart{}, "BookID"); err != nil {
				return err
			}
		}
	}

	return nil
}
