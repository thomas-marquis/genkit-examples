package domain

import (
	"fmt"
	"strings"
	"sync"
)

type GroceryList struct {
	sync.Mutex
	indexedList  map[int]string
	currentIndex int
}

func NewGroceryList() *GroceryList {
	return &GroceryList{
		indexedList:  make(map[int]string),
		currentIndex: 100,
	}
}

func (g *GroceryList) Get() []string {
	l := make([]string, len(g.indexedList))
	i := 0
	for _, item := range g.indexedList {
		l[i] = item
		i++
	}
	return l
}

func (g *GroceryList) Add(item string) int {
	g.Lock()
	defer g.Unlock()
	g.indexedList[g.currentIndex] = item
	g.currentIndex++
	return g.currentIndex - 1
}

func (g *GroceryList) Update(id int, item string) error {
	g.Lock()
	defer g.Unlock()
	if _, ok := g.indexedList[id]; !ok {
		return fmt.Errorf("item with ID %d not found", id)
	}
	g.indexedList[id] = item
	return nil
}

func (g *GroceryList) Delete(id int) error {
	g.Lock()
	defer g.Unlock()
	if _, ok := g.indexedList[id]; !ok {
		return fmt.Errorf("item with ID %d not found", id)
	}
	delete(g.indexedList, id)
	return nil
}

func (g *GroceryList) Len() int {
	return len(g.indexedList)
}

func (g *GroceryList) String() string {
	sb := strings.Builder{}
	for id, item := range g.indexedList {
		sb.WriteString(fmt.Sprintf("- ID=%d ; Item=%s\n", id, item))
	}
	return sb.String()
}
