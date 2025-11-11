package orm

import (
	"genkit-examples/internal/book"

	"gorm.io/datatypes"
)

type Book struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	Title    string `gorm:"not null"`
	Author   string `gorm:"not null"`
	Selected bool   `gorm:"not null;default:false"`
	FileName string
	Metadata datatypes.JSONMap `gorm:"type:jsonb"`
	Status   string
}

// ToDomain converts the ORM entity to domain entity
func (b Book) ToDomain() book.Book {
	var metadata map[string]any
	if b.Metadata == nil {
		metadata = make(map[string]any)
	} else {
		metadata = b.Metadata
	}

	return book.Book{
		ID:       idToString(b.ID),
		Title:    b.Title,
		Metadata: metadata,
		FilePath: b.FileName,
	}
}

// BookFromDomain creates an ORM entity from domain entity
func BookFromDomain(b book.Book) (*Book, error) {
	ormBook := &Book{
		Title:    b.Title,
		Author:   "",
		Metadata: b.Metadata,
		Selected: false,
		FileName: b.FilePath,
		Status:   "",
	}

	if b.ID != "" {
		ormBook.ID = stringToID(b.ID)
	}

	return ormBook, nil
}
