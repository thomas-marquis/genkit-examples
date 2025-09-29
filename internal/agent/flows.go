package agent

import (
	"context"
	"errors"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

type ChatInput struct {
	Question string `json:"question"`
}

type ChatOutput struct {
	Answer string `json:"answer"`
}

//type ragPromptInput struct {
//	Question string `json:"question"`
//}

//func defineChatFlow(g *genkit.Genkit) *core.Flow[ChatInput, ChatOutput, struct{}] {
//	prompt := genkit.DefinePrompt(g, "rag-prompt",
//		ai.WithModelName("mistral/mistral-medium-latest"),
//		ai.WithSystem(`
//You are a multi-purpose assistant. You can answer questions using information provided by the user.
//You don't make up. If the information provided are not sufficient, you can search for a book and suggest the user to purchase it.
//`),
//		ai.WithInputType(ragPromptInput{}),
//		ai.WithPrompt(`Please answer my question: {{question}}`),
//		ai.WithTools(genkit.LookupTool(g, "bookSearchTool")),
//	)
//
//	return genkit.DefineFlow(g, "chat-flow", func(ctx context.Context, input ChatInput) (ChatOutput, error) {
//		// Retrieve documents
//		docs, err := genkit.Retrieve(ctx, g,
//			ai.WithDocs(ai.DocumentFromText(input.Question, nil)),
//			ai.WithRetrieverName("book-retriever"),
//		)
//		if err != nil {
//			return ChatOutput{}, err
//		}
//		if len(docs.Documents) == 0 {
//			return ChatOutput{}, errors.New("no documents found in vector store")
//		}
//
//		// Generate response
//		resp, err := prompt.Execute(ctx,
//			ai.WithInput(ragPromptInput{Question: input.Question}),
//			ai.WithDocs(docs.Documents...),
//		)
//		if err != nil {
//			return ChatOutput{}, err
//		}
//
//		return ChatOutput{
//			Answer: resp.Text(),
//		}, nil
//	})
//}

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
			ai.WithModelName("mistral/mistral-medium-latest"),
			ai.WithSystem(`
You are a multi-purpose assistant. You can answer questions using information provided by the user.
You don't make up. If the information provided are not sufficient, you can search for a book and suggest the user to purchase it.
`),
			ai.WithPrompt("Please answer my question: %s", input.Question),
			ai.WithDocs(docs.Documents...),
			ai.WithTools(genkit.LookupTool(g, "bookSearchTool")),
		)
		if err != nil {
			return ChatOutput{}, err
		}

		return ChatOutput{
			Answer: resp.Text(),
		}, nil
	})
}

func defineSmartFlow(g *genkit.Genkit, tools []ai.ToolRef) *core.Flow[string, string, struct{}] {
	return genkit.DefineFlow(g, "smart-flow", func(ctx context.Context, input string) (string, error) {
		res, err := genkit.Generate(ctx, g,
			ai.WithModelName("mistral/mistral-large-latest"),
			ai.WithSystem(`You are a personal assistant.
Do your best to help the user according to their request.
`),
			ai.WithPrompt(input),
			ai.WithTools(tools...),
		)
		if err != nil {
			return "", err
		}

		return res.Text(), nil
	})
}
