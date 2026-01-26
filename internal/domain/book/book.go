package book

import (
	"fmt"
	"os"

	"github.com/firebase/genkit/go/ai"
	"github.com/timsims/pamphlet"
)

type Book struct {
	ID       string
	Title    string
	Summary  string
	FilePath string
	Metadata map[string]any
}

func (b *Book) Parse() ([]*ai.Document, error) {
	conv := makeMarkdownConverter()
	splitter := makeTextSplitter()

	file, _ := os.Open(b.FilePath)
	parser, err := pamphlet.OpenFile(file)
	if err != nil {
		return nil, err
	}

	book := parser.GetBook()
	chunks := make([]*ai.Document, 0, len(book.Chapters))
	for i, chap := range book.Chapters {
		// Get XHTML content
		chapContent, err := chap.GetContent()
		if err != nil {
			return nil, err
		}

		// Convert to Markdown
		markdown, err := conv.ConvertString(chapContent)
		if err != nil {
			return nil, fmt.Errorf("failed to convert chapter %d to markdown: %w", i+1, err)
		}

		if markdown == "" {
			continue
		}

		// Chunk in smaller parts
		docs, err := splitDocument(splitter, markdown)
		if err != nil {
			return nil, err
		}

		chunks = append(chunks, docs...)
	}
	return chunks, nil
}
