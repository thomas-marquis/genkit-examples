package qna

import (
	"fmt"
	"genkit-examples/internal/domain/book"

	"github.com/firebase/genkit/go/ai"
)

const (
	searchBookToolName = "searchBook"
)

type searchBookInput struct {
	SearchQuery string `json:"search_query" jsonschema_description:"A concise search query for the books search engine"`
}

func (a *Agent) searchBookToolHandler(ctx *ai.ToolContext, input searchBookInput) ([]book.Book, error) {
	results, err := a.bookRepository.Search(input.SearchQuery)
	if err != nil {
		return nil, fmt.Errorf("tool %s failing to search for books: %w", searchBookToolName, err)
	}
	return results, nil
}
