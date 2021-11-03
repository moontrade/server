package nosql_test

import (
	"github.com/moontrade/server/storage/nosql"
	"testing"
)

func TestSchema(t *testing.T) {
	dbSchema := &dbSchema{}
	schema, err := nosql.Load(dbSchema)
	if err != nil {
		t.Fatal(err)
	}
	_ = schema
}

type dbSchema struct {
	Orders Orders

	Items struct {
		nosql.Collection
		Price nosql.Float64 `@:"price"`
	}

	Markets struct {
		nosql.Collection
		Num   nosql.Int64  `@:"num" unique:"true"`
		Key   nosql.String `@:"key" unique:"true"`
		Name  nosql.String `@:"name"`
		Names struct {     // Composite index
			First nosql.String `@:""`
			Last  nosql.String `@:""`
		}
	}
}

type Order struct {
	ID    uint64 `json:"id"`
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
	Num       nosql.UniqueInt64  `@:"num"`
	Key       nosql.UniqueString `@:"key"`
	Price     nosql.Float64      `@:"price"`
	FirstName nosql.String       `@:"name.first"`
}
