package agent

import (
	"context"
	"fmt"
	"genkit-examples/internal/book"
	"genkit-examples/internal/vectorstore"
	"log"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/mcp"
)

type Agent struct {
	g              *genkit.Genkit
	ctx            context.Context
	vecStore       *vectorstore.VectorStore
	chatFlow       *core.Flow[ChatInput, ChatOutput, struct{}]
	bookRepository book.Repository
	notionMcp      *mcp.GenkitMCPClient
}

func New(ctx context.Context, g *genkit.Genkit, vecStore *vectorstore.VectorStore, bookRepository book.Repository, notionKey string) *Agent {
	genkit.DefineRetriever(g, "book-retriever", &ai.RetrieverOptions{}, vecStore.MakeRetrieverHandler(g))
	defineBookSearchTool(g, bookRepository)
	mcpClient, err := defineNotionMcp(notionKey)
	if err != nil {
		log.Fatal(err)
	}

	tools, err := mcpClient.GetActiveTools(ctx, g)
	if err != nil {
		log.Fatal(err)
	}
	toolRefs := make([]ai.ToolRef, len(tools))
	for i, t := range tools {
		toolRefs[i] = t
	}
	res, err := defineSmartFlow(g, toolRefs).Run(ctx, "Créee une page Notion dans le projet Genkit dans laquelle tu fera une introduction à la concurrence en go. Si la page existe déjà, met simplement à jour son contenu")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)

	return &Agent{
		g:              g,
		ctx:            ctx,
		vecStore:       vecStore,
		chatFlow:       defineChatFlow(g),
		bookRepository: bookRepository,
		notionMcp:      mcpClient,
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
