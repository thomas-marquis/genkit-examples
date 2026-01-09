package infrastructure

import (
	"context"
	"errors"
	"genkit-examples/internal/domain"
)

var (
	ErrListNotFound = errors.New("grocery list not found")
)

type groceryListRepoImpl struct {
	gl *domain.GroceryList
}

func NewGroceryListRepositoryImpl() domain.GroceryListRepository {
	return &groceryListRepoImpl{}
}

func (r *groceryListRepoImpl) Save(ctx context.Context, list *domain.GroceryList) error {
	r.gl = list
	return nil
}

func (r *groceryListRepoImpl) Load(ctx context.Context) (*domain.GroceryList, error) {
	if r.gl != nil {
		return r.gl, nil
	}
	return nil, ErrListNotFound
}
