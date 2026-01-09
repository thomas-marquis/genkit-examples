package domain

import "context"

type GroceryListRepository interface {
	Save(ctx context.Context, list *GroceryList) error
	Load(ctx context.Context) (*GroceryList, error)
}
