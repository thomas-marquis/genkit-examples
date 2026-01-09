package agent

import (
	"context"
	"errors"
	"fmt"
	"genkit-examples/internal/domain"
	"sync"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	mistralclient "github.com/thomas-marquis/mistral-client/mistral"
)

var (
	errMenuPlanner       = errors.New("menu planner error")
	menuPlannerNbWorkers = 5
)

type menuPlannerInput struct {
	NbMeals    uint8  `jsonschema_description:"The number of meals to create"`
	Constraint string `jsonschema_description:"The constraint to respect"`
	Language   string `jsonschema_description:"The language of the recipes (english, french, ...). Default to English."`
}

type menuPlannerOutput struct {
	Recipes []domain.Recipe `jsonschema_description:"The list of recipes created"`
}

func (a *MenuAgent) newMenuPlannerFlow(
	g *genkit.Genkit,
	createRecipe *core.Flow[recipeCreatorInput, domain.Recipe, struct{}],
	groceryListManager *core.Flow[groceryListManagerInput, string, struct{}],
) *core.Flow[menuPlannerInput, menuPlannerOutput, struct{}] {
	prompt := genkit.DefinePrompt(g, "menuPrompt",
		ai.WithModelName("mistral/mistral-small-latest"),
		ai.WithConfig(mistralclient.CompletionConfig{
			Temperature: 0.7,
		}),
		ai.WithSystem(`You are a menu planner for an individual. Just give a list of courses without any details according the the constraints given by the user.
Don't create the full recipe, just a short description of the meal. Examples:
- Grilled Salmon with Quinoa and Steamed Broccoli
- Chickpea and Spinach Curry with Brown Rice
- Stuffed Bell Peppers with Lean Turkey and Quinoa

Respect the number of meal to create given by the user.
Use this language to answer: {{language}}
'
`),
		ai.WithPrompt(`Create a {{nb}}-meals menu that respect this constraint: {{constraint}}`),
		ai.WithOutputType([]string{}))

	type groceryListStepInput struct {
		FlowInput groceryListManagerInput
		Recipe    domain.Recipe
	}

	return genkit.DefineFlow(g, "menuPlanner", func(ctx context.Context, in menuPlannerInput) (menuPlannerOutput, error) {
		menus, err := genkit.Run(ctx, "createMenus", func() ([]string, error) {
			menuRes, err := prompt.Execute(ctx,
				ai.WithInput(map[string]any{
					"nb":         in.NbMeals,
					"constraint": in.Constraint,
					"language":   in.Language,
				}))
			if err != nil {
				return nil, err
			}
			var menus []string
			if err := menuRes.Output(&menus); err != nil {
				return nil, err
			}
			return menus, nil
		})
		if err != nil {
			return menuPlannerOutput{}, errors.Join(errMenuPlanner, errors.New("createMenus"), err)
		}

		return genkit.Run(ctx, "createRecipesAndUpdateGroceryList", func() (menuPlannerOutput, error) {
			recipes := make([]domain.Recipe, 0, len(menus))
			createRecipeChan := make(chan recipeCreatorInput)
			updateGroceryListChan := make(chan groceryListStepInput, 10)
			outChan := make(chan domain.Recipe)
			done := make(chan struct{})
			closeDone := sync.OnceFunc(func() {
				close(done)
			})

			cancelableCtx, cancel := context.WithCancelCause(ctx)
			defer cancel(context.Canceled)

			go pipelineStage(cancel, createRecipeChan, updateGroceryListChan, func(input recipeCreatorInput) (groceryListStepInput, error) {
				fmt.Printf("Creating recipe for menu: '%s'\n", input.Instructions)
				recipe, err := createRecipe.Run(cancelableCtx, input)
				if err != nil {
					return groceryListStepInput{}, errors.Join(errMenuPlanner, err)
				}
				return groceryListStepInput{
					FlowInput: groceryListManagerInput{
						Items:    recipe.Ingredients,
						Language: input.Language,
					},
					Recipe: recipe,
				}, nil
			})

			go pipelineStage(cancel, updateGroceryListChan, outChan, func(input groceryListStepInput) (domain.Recipe, error) {
				_, err = groceryListManager.Run(cancelableCtx, input.FlowInput)
				if err != nil {
					return domain.Recipe{}, errors.Join(errMenuPlanner, err)
				}
				return input.Recipe, nil
			})

			go func() {
				for recipe := range outChan {
					fmt.Println("Recipe created:", recipe.String())
					recipes = append(recipes, recipe)
				}
				closeDone()
			}()

			go func() {
				<-cancelableCtx.Done()
				closeDone()

			}()

			fmt.Println("Menus to create:")
			for _, m := range menus {
				fmt.Println(m)

				select {
				case <-cancelableCtx.Done():
					return menuPlannerOutput{}, cancelableCtx.Err()
				case createRecipeChan <- recipeCreatorInput{
					Instructions: m,
					Language:     in.Language,
				}:
				}
			}
			close(createRecipeChan)
			<-done

			return menuPlannerOutput{
				Recipes: recipes,
			}, nil
		})
	})
}

func pipelineStage[IN any, OUT any](cancel context.CancelCauseFunc, in <-chan IN, out chan<- OUT, handler func(IN) (OUT, error)) {
	defer close(out)
	for data := range in {
		o, err := handler(data)
		if err != nil {
			cancel(err)
			return
		}
		out <- o
	}
}
