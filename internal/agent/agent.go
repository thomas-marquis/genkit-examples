package agent

import (
	"context"
	"genkit-examples/internal/domain"

	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

type Input struct {
	Instruction string `jsonschema_description:"The instruction to follow: how many meals to create and what constraint to respect"`
}

type Output struct {
	Recipes     []domain.Recipe `jsonschema_description:"The list of recipes created"`
	GroceryList []string        `jsonschema_description:"The current grocery list"`
}

type MenuAgent struct {
	glRepo domain.GroceryListRepository

	preprocessor *core.Flow[string, preprocessorOutput, struct{}]
	menuPlanner  *core.Flow[menuPlannerInput, menuPlannerOutput, struct{}]
	mainFlow     *core.Flow[Input, Output, struct{}]
}

func NewMenuAgent(g *genkit.Genkit, glRepo domain.GroceryListRepository) *MenuAgent {
	a := &MenuAgent{glRepo: glRepo}

	createRecipe := a.newRecipeCreatorFlow(g)
	groceryListManager := a.newGroceryListManagerFlow(g, a.glRepo)
	a.menuPlanner = a.newMenuPlannerFlow(g, createRecipe, groceryListManager)
	a.preprocessor = a.newPreprocessor(g)

	a.mainFlow = genkit.DefineFlow(g, "mainAgentFlow", a.handler)

	return a
}

func (a *MenuAgent) Flow() *core.Flow[Input, Output, struct{}] {
	return a.mainFlow
}

func (a *MenuAgent) handler(ctx context.Context, in Input) (Output, error) {
	prep, err := a.preprocessor.Run(ctx, in.Instruction)
	if err != nil {
		return Output{}, err
	}

	groceryList := domain.NewGroceryList()
	if err := a.glRepo.Save(ctx, groceryList); err != nil {
		return Output{}, err
	}

	res, err := a.menuPlanner.Run(ctx, prep.menuPlannerInput)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Recipes:     res.Recipes,
		GroceryList: groceryList.Get(),
	}, nil
}
