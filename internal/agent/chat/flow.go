package chat

import (
	"context"
	"errors"
	"genkit-examples/internal/book"
	"net/http"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

const (
	// Name is the Flow's name
	Name = "chat"

	maxTurns     = 10
	systemPrompt = `
You are a general assistant who use books to answer questions about various topics.
You are not an expert in any domain, your job consists in summarizing book content (if some are available) or suggesting books to get.
Don't make up answers. If the response is not provided in the provided book extracts, try to search for a book online and suggest some reference to the user.

Important: use the provided tools as many time you need.

## When you have enough information:
Use ONLY the provided documents to answer the user's question.
If it isn't enough, you can also search for books online and suggest some reference to the user.
Your response must feel natural, so don't make any reference to the provided documents 
(those documents are provided automatically, the user don't have to know about them').'

## When you don't have enough information:
Search for books with your tools and suggest some of them to the user.
If available, provide links to get each book.
Don't make up a book that doesn't exist.
Never suggest without searching for books.
`
	defaultLLMModelName       = "mistral/mistral-small-latest"
	defaultEmbeddingModelName = "mistral/mistral-embed"
)

type Flow struct {
	g                  *genkit.Genkit
	flow               *core.Flow[Input, *Output, struct{}]
	bookRepository     book.Repository
	llmModelName       string
	embeddingModelName string
	notionApiKey       string
}

func New(ctx context.Context, g *genkit.Genkit, bookRepository book.Repository, options ...Option) (*Flow, error) {
	f := &Flow{g: g, bookRepository: bookRepository}

	for _, opt := range options {
		opt(f)
	}

	if f.llmModelName == "" {
		f.llmModelName = defaultLLMModelName
	}
	if f.embeddingModelName == "" {
		f.embeddingModelName = defaultEmbeddingModelName
	}

	f.flow = genkit.DefineFlow(g, Name, f.ragHandler)

	genkit.DefineTool(g, searchBookToolName,
		"Search for a book calling an external API. Use this tool when the documents provided by the user are not enough to answer the user's question.",
		f.searchBookToolHandler,
	)

	genkit.DefineRetriever(g, bookRetrieverName, &ai.RetrieverOptions{}, f.bookRetrieverHandler)

	if err := f.setupMCPClient(ctx); err != nil {
		return nil, err
	}

	return f, nil
}

func (f *Flow) ToHandler() http.HandlerFunc {
	return genkit.Handler(f.flow)
}

func (f *Flow) ragHandler(ctx context.Context, input Input) (*Output, error) {
	// 1. Retrieve documents
	docs, err := genkit.Retrieve(ctx, f.g,
		ai.WithDocs(ai.DocumentFromText(input.Question, nil)),
		ai.WithRetrieverName(bookRetrieverName),
	)
	if err != nil {
		return nil, err
	}
	if len(docs.Documents) == 0 {
		return nil, errors.New("no documents found in vector store")
	}

	// 2. Generate response
	resp, err := genkit.Generate(ctx, f.g,
		ai.WithModelName(f.llmModelName),
		ai.WithSystem(systemPrompt),
		ai.WithPrompt("Please answer my question: %s", input.Question),
		ai.WithDocs(docs.Documents...),
		ai.WithTools(
			genkit.LookupTool(f.g, searchBookToolName),
		),
		ai.WithMaxTurns(maxTurns),
	)
	if err != nil {
		return nil, err
	}

	return &Output{
		Answer: resp.Text(),
	}, nil
}
