package vectorstore

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/pgvector/pgvector-go"
)

type Chunk struct {
	ID        uint            `gorm:"primaryKey;autoIncrement"`
	Embedding pgvector.Vector `gorm:"type:vector(1024);not null"`
	Content   string          `gorm:"not null;uniqueIndex:idx_book_content"`
	BookID    string          `gorm:"not null;uniqueIndex:idx_book_content"`
}

func (c *Chunk) ToDocument() *ai.Document {
	return ai.DocumentFromText(c.Content, nil)
}

func toDocuments(chunks []*Chunk) []*ai.Document {
	docs := make([]*ai.Document, 0, len(chunks))
	for _, chunk := range chunks {
		docs = append(docs, chunk.ToDocument())
	}
	return docs
}
