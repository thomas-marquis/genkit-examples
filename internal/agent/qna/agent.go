package qna

import (
	"context"
	"errors"
	"genkit-examples/internal/domain/book"
	"log"
	"net/http"
	"os"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

var (
	logger = log.New(os.Stdout, "QnA", log.LstdFlags)
)

const (
	FlowName = "qna"

	systemPrompt = `
You are a general assistant who use books to answer questions about various topics.
You are not an expert in any domain, your job consists in summarizing book content (if some are available) or suggesting books to get.
Don't make up answers. If the response is not provided in the provided book extracts, try to search for a book online and suggest some reference to the user.
In such a cas, DON'T TRY TO ANSWER THE USER, just suggest books to get.

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
)

type Input struct {
	Question string `json:"question"`
}

type Output struct {
	Answer string `json:"answer"`
}

type Agent struct {
	mainFlow       *core.Flow[Input, Output, struct{}]
	g              *genkit.Genkit
	bookRepository book.Repository
	todoistApiKey  string
}

func New(g *genkit.Genkit, repo book.Repository, todoistApiKey string) *Agent {
	a := &Agent{
		todoistApiKey:  todoistApiKey,
		bookRepository: repo,
		g:              g,
	}

	a.mainFlow = genkit.DefineFlow(g, FlowName, a.mainFlowHandler)

	genkit.DefineRetriever(g, bookRetrieverName, &ai.RetrieverOptions{}, a.bookRetrieverHandler)

	genkit.DefineTool(g, searchBookToolName,
		"Search for a book calling an external API.",
		a.searchBookToolHandler,
	)

	return a
}

func (a *Agent) Flow() *core.Flow[Input, Output, struct{}] {
	return a.mainFlow
}

func (a *Agent) ToHandler() http.HandlerFunc {
	return genkit.Handler(a.mainFlow)
}

func (a *Agent) mainFlowHandler(ctx context.Context, in Input) (Output, error) {
	// 1. Retrieve documents
	docs, err := genkit.Retrieve(ctx, a.g,
		ai.WithDocs(ai.DocumentFromText(in.Question, nil)),
		ai.WithRetrieverName(bookRetrieverName),
	)

	if err != nil {
		return Output{}, err
	}
	if len(docs.Documents) == 0 {
		return Output{}, errors.New("no documents found")
	}

	// 2. Generate response
	resp, err := genkit.Generate(ctx, a.g,
		ai.WithModelName("mistral/mistral-large-latest"),
		ai.WithSystem(systemPrompt),
		ai.WithPrompt("Please answer my question: %s", in.Question),
		ai.WithDocs(docs.Documents...),
		ai.WithTools(
			genkit.LookupTool(a.g, searchBookToolName),
		),
		ai.WithMaxTurns(10),
	)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Answer: resp.Text(),
	}, nil
}
