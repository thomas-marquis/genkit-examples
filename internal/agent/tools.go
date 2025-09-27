package agent

import (
	"fmt"
	"genkit-examples/internal/book"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type SearchBookInput struct {
	SearchQuery string `json:"search_query" jsonschema_description:"A concise search query for the books search engine"`
}

func defineBookSearchTool(g *genkit.Genkit, client book.Repository) ai.ToolRef {
	return genkit.DefineTool(g, "bookSearchTool",
		"Search for a book calling an external API. Use this tool when the documents provided by the user are not enough to answer the user's question.",
		func(ctx *ai.ToolContext, question SearchBookInput) ([]book.Book, error) {
			results, err := client.Search(question.SearchQuery)
			if err != nil {
				return nil, fmt.Errorf("tool search-book-tool failing to search for books: %w", err)
			}
			return results, nil
		})
}
