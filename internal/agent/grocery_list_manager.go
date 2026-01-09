package agent

import (
	"context"
	"errors"
	"fmt"
	"genkit-examples/internal/domain"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	mistralclient "github.com/thomas-marquis/mistral-client/mistral"
)

var (
	errGroceryListManager = errors.New("grocery list manager error")
)

type groceryListAddInput struct {
	Item string `jsonschema_description:"The item to add to the list and the quantity indication"`
}

type groceryListDeleteInput struct {
	ID int `jsonschema_description:"The ID of the item to remove from the list"`
}

type groceryListUpdateInput struct {
	ID          int    `jsonschema_description:"The ID of the item to update in the list"`
	UpdatedItem string `jsonschema_description:"The new item to replace the old one with the updated label and/or quantity indication"`
}

type groceryListManagerInput struct {
	Items    []string `jsonschema_description:"The list of items to add to the grocery list"`
	Language string   `jsonschema_description:"The language of the item (english, french, ...). Default to English."`
}

func (a *MenuAgent) newGroceryListManagerFlow(g *genkit.Genkit, repo domain.GroceryListRepository) *core.Flow[groceryListManagerInput, string, struct{}] {
	groceryListGet := genkit.DefineTool(g, "groceryListGet", "get the current content of the list", func(ctx *ai.ToolContext, input any) (string, error) {
		fmt.Println("groceryListGet")
		groceryList, err := repo.Load(ctx)
		if err != nil {
			return "", errors.Join(errGroceryListManager, errors.New("groceryListGet tool"), err)
		}
		if groceryList.Len() == 0 {
			return "The list is empty", nil
		}
		return groceryList.String(), nil
	})

	groceryListAdd := genkit.DefineTool(g, "groceryListAdd", "add an item to the list", func(ctx *ai.ToolContext, input groceryListAddInput) (string, error) {
		fmt.Printf("groceryListAdd: %s\n", input.Item)
		groceryList, err := repo.Load(ctx)
		if err != nil {
			return "", errors.Join(errGroceryListManager, errors.New("groceryListAdd tool"), err)
		}
		groceryList.Add(input.Item)
		if err := repo.Save(ctx, groceryList); err != nil {
			return "", errors.Join(errGroceryListManager, errors.New("groceryListAdd tool"), err)
		}
		return "ok", nil
	})

	groceryListDelete := genkit.DefineTool(g, "groceryListDelete", "remove an item from the grocery list", func(ctx *ai.ToolContext, input groceryListDeleteInput) (string, error) {
		fmt.Printf("groceryListDelete: %d\n", input.ID)
		groceryList, err := repo.Load(ctx)
		if err != nil {
			return "", errors.Join(errGroceryListManager, errors.New("groceryListDelete tool"), err)
		}
		if err := groceryList.Delete(input.ID); err != nil {
			return "", errors.Join(errGroceryListManager, errors.New("groceryListDelete tool"), err)
		}
		if err := repo.Save(ctx, groceryList); err != nil {
			return "", errors.Join(errGroceryListManager, errors.New("groceryListDelete tool"), err)
		}
		return "ok", nil
	})

	groceryListUpdate := genkit.DefineTool(g, "groceryListUpdate", "update an item in the grocery list", func(ctx *ai.ToolContext, input groceryListUpdateInput) (string, error) {
		fmt.Printf("groceryListUpdate: %d %s\n", input.ID, input.UpdatedItem)
		groceryList, err := repo.Load(ctx)
		if err != nil {
			return "", errors.Join(errGroceryListManager, errors.New("groceryListUpdate tool"), err)
		}
		if err := groceryList.Update(input.ID, input.UpdatedItem); err != nil {
			return "", errors.Join(errGroceryListManager, errors.New("groceryListUpdate tool"), err)
		}
		if err := repo.Save(ctx, groceryList); err != nil {
			return "", errors.Join(errGroceryListManager, errors.New("groceryListUpdate tool"), err)
		}
		return fmt.Sprintf("UpdatedItem %d updated in the list", input.ID), nil
	})

	prompt := genkit.DefinePrompt(g, "groceryListPrompt",
		ai.WithSystem(`You are a grocery list manager. Your role is to keep it as clean and coherent as possible. 
Instructions:
- Use this language: {{language}}
- Avoid duplication: if an item already exists in the list, update it instead of adding a new one.
- Get the list content times to times
- Ensure you don't forget any item on the list.
- Use this format for list items: "<label>, <quantity>"
- Avoid recipe-related comment in the list item label, keep it simple`),
		ai.WithPrompt(`Add or update the grocery list with the following items:
{{#each items}}- {{this}}{{/each}}
`),
		ai.WithModelName("mistral/mistral-medium-latest"),
		ai.WithConfig(mistralclient.CompletionConfig{
			Temperature:       0.1,
			ParallelToolCalls: true,
		}),
		ai.WithTools(
			groceryListGet,
			groceryListAdd,
			groceryListUpdate,
			groceryListDelete,
		),
	)

	return genkit.DefineFlow(g, "groceryListManagerFlow", func(ctx context.Context, in groceryListManagerInput) (string, error) {
		res, err := prompt.Execute(ctx,
			ai.WithInput(map[string]interface{}{
				"items":    in.Items,
				"language": in.Language,
			}),
			ai.WithReturnToolRequests(false),
			ai.WithMaxTurns(25),
		)
		if err != nil {
			return "", errors.Join(errGroceryListManager, err)
		}
		return res.Text(), nil
	})
}
