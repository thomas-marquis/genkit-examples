package book

type Repository interface {
	Search(query string) ([]Book, error)
}
