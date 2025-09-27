package agent

import (
	"context"
	"genkit-examples/internal/book"
	"genkit-examples/internal/vectorstore"
	"log"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

type Agent struct {
	g              *genkit.Genkit
	ctx            context.Context
	vecStore       *vectorstore.VectorStore
	chatFlow       *core.Flow[ChatInput, ChatOutput, struct{}]
	bookRepository book.Repository
}

func New(ctx context.Context, g *genkit.Genkit, vecStore *vectorstore.VectorStore, bookRepository book.Repository) *Agent {
	genkit.DefineRetriever(g, "book-retriever", &ai.RetrieverOptions{}, vecStore.MakeRetrieverHandler(g))
	defineBookSearchTool(g, bookRepository)
	defineRagPrompt(g)

	return &Agent{
		g:              g,
		ctx:            ctx,
		vecStore:       vecStore,
		chatFlow:       defineChatFlow(g),
		bookRepository: bookRepository,
	}
}

func (a *Agent) ChatFlow() *core.Flow[ChatInput, ChatOutput, struct{}] {
	return a.chatFlow
}

func (a *Agent) Ask(question string) (string, error) {
	ctx := context.Background()
	res, err := a.chatFlow.Run(ctx, ChatInput{Question: question})
	if err != nil {
		return "", err
	}
	return res.Answer, nil
}

func (a *Agent) SimpleTextGeneration() {
	res, err := genkit.Generate(a.ctx, a.g,
		ai.WithModelName("mistral/mistral-small-latest"),
		ai.WithPrompt("Hello, how are you?"),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(res.Text())
}
