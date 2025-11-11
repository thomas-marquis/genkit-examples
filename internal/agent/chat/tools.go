package chat

import (
	"fmt"
	"genkit-examples/internal/book"

	"github.com/firebase/genkit/go/ai"
)

const (
	searchBookToolName = "chatSearchBook"
	searchBookToolDesc = `
Search for books using an external API. This tool allows you to search by keyword in a big book catalogue online.
Call this tool each time you want to suggest a book to the user.
Call it multiple times with different keywords if needed (to get more books or if results are not relevant).
`
)

func (f *Flow) searchBookToolHandler(ctx *ai.ToolContext, input SearchBookInput) ([]book.Book, error) {
	results, err := f.bookRepository.Search(input.SearchQuery)
	if err != nil {
		return nil, fmt.Errorf("tool %s failing to search for books: %w", searchBookToolName, err)
	}
	return results, nil
}
