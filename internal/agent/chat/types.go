package chat

import (
	"log"
	"os"
)

var (
	logger = log.New(os.Stdout, "chat", log.LstdFlags)
)

type Input struct {
	Question string `json:"question"`
}

type Output struct {
	Answer string `json:"answer"`
}

type SearchBookInput struct {
	SearchQuery string `json:"search_query" jsonschema_description:"A concise search query for the books search engine"`
}
