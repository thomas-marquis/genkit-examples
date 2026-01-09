package agent

import (
	"context"
	"errors"
	"genkit-examples/internal/domain"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	mistralclient "github.com/thomas-marquis/mistral-client/mistral"
)

var (
	errRecipeCreator = errors.New("recipe creator error")
)

type recipeCreatorInput struct {
	Instructions string `jsonschema_description:"The instructions to follow to create the recipe as well as the difficulty level, expected duration, ..."`
	Language     string `jsonschema_description:"The language of the recipe (en, fr, ...). Default to English."`
}

func (a *MenuAgent) newRecipeCreatorFlow(g *genkit.Genkit) *core.Flow[recipeCreatorInput, domain.Recipe, struct{}] {
	prompt := genkit.DefinePrompt(g, "recipePrompt",
		ai.WithSystem(`You are a professional and experienced chef. You are creative and pragmatic.
Answer in this language: {{language}}`),
		ai.WithOutputType(domain.Recipe{}),
		ai.WithConfig(mistralclient.CompletionConfig{
			Temperature: 0.7,
		}),
		ai.WithModelName("mistral/mistral-small-latest"),
		ai.WithPrompt(`Create a complete recipe according to the following instructions:
{{instructions}}`))

	return genkit.DefineFlow(g, "recipeCreator", func(ctx context.Context, in recipeCreatorInput) (domain.Recipe, error) {
		var recipe domain.Recipe
		res, err := prompt.Execute(ctx,
			ai.WithInput(map[string]interface{}{
				"instructions": in.Instructions,
				"language":     in.Language,
			}))
		if err != nil {
			return domain.Recipe{}, errors.Join(errRecipeCreator, err)
		}
		if err := res.Output(&recipe); err != nil {
			return domain.Recipe{}, errors.Join(errRecipeCreator, err)
		}

		return recipe, nil
	})
}
