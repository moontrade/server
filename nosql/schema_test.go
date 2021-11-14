package nosql_test

import (
	"github.com/moontrade/mdbx-go"
	"github.com/moontrade/server/nosql"
	"testing"
)

func TestSchema(t *testing.T) {
	store, err := nosql.Open(&nosql.Config{
		Path:  "./testdata",
		Flags: nosql.DefaultDurable,
		Mode:  0755,
	})
	if err != nil {
		t.Fatal(err)
	}

	var (
		schema    = &Schema{}
		nschema   *nosql.Schema
		evolution *nosql.evolution
	)
	if nschema, err = nosql.ParseSchemaWithUID("@", schema); err != nil {
		t.Fatal(err)
	}
	if evolution, err = store.Hydrate(nschema); err != nil {
		t.Fatal(err)
	}
	_, err = evolution.Apply()
}

type Schema struct {
	*nosql.Schema

	Orders struct {
		_ Order
		nosql.Collection
		Num       nosql.Int64Unique  `@:"num"`
		Key       nosql.StringUnique `@:"key"`
		Price     nosql.Float64      `@:"price"`
		FirstName nosql.String       `@:"name.first"`
	}

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

type Schema2 struct {
	Orders Orders

	Items struct {
		nosql.Collection
		Price  nosql.Float64 `@:"price"`
		Price2 nosql.Float64 `@:"price2"`
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
	ID    uint64  `json:"ID"`
	Num   uint64  `json:"num"`
	Key   string  `json:"key"`
	Price float64 `json:"price"`
	Name  struct {
		First string `json:"first"`
		Last  string `json:"last"`
	} `json:"name"`
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
