package nosql_test

import (
	"github.com/moontrade/mdbx-go"
	"github.com/moontrade/server/storage/nosql"
	"testing"
)

var (
	schema = &Schema{}
	Items  = &schema.Items
)

func TestSchema(t *testing.T) {
	s, err := nosql.ParseSchema(schema)

	if err != nil {
		t.Fatal(err)
	}

	_ = s
}

type Schema struct {
	Orders Orders

	Items struct {
		nosql.Collection
		Price nosql.Float64 `@:"price"`
	}

	Markets struct {
		nosql.Collection
		Num   nosql.Int64  `@:"num"`
		Key   nosql.String `@:"key"`
		Name  nosql.String `@:"name"`
		Names struct {
			First nosql.String `sort:"ASC"`
			Last  nosql.String `sort:"DESC"`
		} `@:"[name.first,age,children.0]"` // Composite index
	}
}

type Order struct {
	ID    uint64 `json:"ID"`
	Num   uint64 `json:"num"`
	Key   string `json:"key"`
	Price float64
	Name  struct {
		First string
		Last  string
	}
}

type Orders struct {
	_ Order
	nosql.Collection
	Num       nosql.Int64Unique  `@:"num"`
	Key       nosql.StringUnique `@:"key"`
	Price     nosql.Float64      `@:"price"`
	FirstName nosql.String       `@:"name.first"`
}

func (s *Orders) UpdateWith(tx *mdbx.Tx, id nosql.DocID, data *Order) error {
	return nil
}
