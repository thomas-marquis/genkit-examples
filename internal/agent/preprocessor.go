package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

var (
	errPreprocessor = errors.New("preprocessor error")
)

type preprocessorOutput struct {
	menuPlannerInput
	errorMessage string `jsonschema_description:"The error message if any. Leave empty if no error."`
}

func (a *MenuAgent) newPreprocessor(g *genkit.Genkit) *core.Flow[string, preprocessorOutput, struct{}] {
	prompt := genkit.DefinePrompt(g, "preprocessorPrompt",
		ai.WithModelName("mistral/mistral-small-latest"),
		ai.WithSystem(`
You work with a menu planner. you are responsible get the users request and give structured instructions to the menu planner.
Don't plan th menu by yourself.

Respect these rules:
- The user must ask about planning one or many meal. Reject any other demands
- Answer the user in the same language as he/she wrote
'
`),
		ai.WithPrompt(`{{input}}`),
		ai.WithOutputType(menuPlannerInput{}))

	return genkit.DefineFlow(g, "preprocessor", func(ctx context.Context, in string) (preprocessorOutput, error) {
		res, err := prompt.Execute(ctx,
			ai.WithInput(map[string]any{
				"input": in,
			}))
		if err != nil {
			return preprocessorOutput{}, errors.Join(errPreprocessor, err)
		}
		var out preprocessorOutput
		if err := res.Output(&out); err != nil {
			return preprocessorOutput{}, errors.Join(errPreprocessor, err)
		}
		if out.errorMessage != "" {
			return preprocessorOutput{}, errors.Join(errPreprocessor, fmt.Errorf(
				"an error occurred during the preprocessing: %w",
				errors.New(out.errorMessage)))
		}
		return out, nil
	})
}
