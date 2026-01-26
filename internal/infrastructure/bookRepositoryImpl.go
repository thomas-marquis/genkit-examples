package infrastructure

import (
	"genkit-examples/internal/domain/book"
	"net/http"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type bookRepositoryImpl struct {
	httpClient *http.Client
	maxResults int
	db         *gorm.DB
}

// NewBookRepositoryImpl creates a Google Books API and PG Vector client.
// maxResults controls how many results to return (1-40). Values outside this range are clamped.
func NewBookRepositoryImpl(maxResults int, pgConnection string) (book.Repository, error) {
	if maxResults < 1 {
		maxResults = 10
	}
	if maxResults > 40 {
		maxResults = 40
	}

	db, err := gorm.Open(postgres.Open(pgConnection), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	r := &bookRepositoryImpl{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		maxResults: maxResults,
		db:         db,
	}

	if err := r.migrate(); err != nil {
		return nil, err
	}

	return r, nil
}
