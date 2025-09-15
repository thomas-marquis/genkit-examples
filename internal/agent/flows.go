package agent

import (
	"context"
	"errors"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

type ChatInput struct {
	Question string
}

type ChatOutput struct {
	Answer string
}

func defineChatFlow(g *genkit.Genkit) *core.Flow[ChatInput, ChatOutput, struct{}] {
	return genkit.DefineFlow(g, "chat-flow", func(ctx context.Context, input ChatInput) (ChatOutput, error) {
		// Retrieve documents
		docs, err := genkit.Retrieve(ctx, g,
			ai.WithDocs(ai.DocumentFromText(input.Question, nil)),
			ai.WithRetrieverName("book-retriever"),
		)
		if err != nil {
			return ChatOutput{}, err
		}
		if len(docs.Documents) == 0 {
			return ChatOutput{}, errors.New("no documents found in vector store")
		}

		// Generate response
		resp, err := genkit.Generate(ctx, g,
			ai.WithModelName("mistral/mistral-small-latest"),
			ai.WithSystem("you're a useful assistant, use only the provided documents to answer if they're relevant. Don't make up."),
			ai.WithPrompt(input.Question),
			ai.WithDocs(docs.Documents...),
		)
		if err != nil {
			return ChatOutput{}, err
		}

		return ChatOutput{
			Answer: resp.Text(),
		}, nil
	})
}
