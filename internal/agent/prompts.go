package agent

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type ragPromptInput struct {
	Question string `json:"question"`
}

func defineRagPrompt(g *genkit.Genkit) {
	genkit.DefinePrompt(g, "rag-prompt",
		ai.WithModelName("mistral/mistral-medium-latest"),
		ai.WithSystem(`
You are a multi-purpose assistant. You can answer questions using information provided by the user.
You don't make up. If the information provided are not sufficient, you can search for a book and suggest the user to purchase it.
`),
		ai.WithInputType(ragPromptInput{}),
		ai.WithPrompt(`Please answer my question: {{question}}`),
		ai.WithTools(genkit.LookupTool(g, "bookSearchTool")),
	)
}
