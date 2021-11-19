package nosql_test

import (
	"context"
	"fmt"
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
		mySchema  = &MySchema{}
		mySchema2 = &Schema2{}
		schema    *nosql.Schema
		progress  <-chan nosql.EvolutionProgress
	)
	if schema, err = nosql.ParseSchemaWithUID("@", mySchema); err != nil {
		t.Fatal(err)
	}
	if progress, err = store.Hydrate(context.Background(), schema); err != nil {
		t.Fatal(err)
	}
	wait(progress)

	if schema, err = nosql.ParseSchemaWithUID("@", mySchema2); err != nil {
		t.Fatal(err)
	}
	if progress, err = store.Hydrate(context.Background(), schema); err != nil {
		t.Fatal(err)
	}
	wait(progress)
}

func wait(progress <-chan nosql.EvolutionProgress) {
	if progress == nil {
		fmt.Println("schema does not require an evolution")
		return
	}
LOOP:
	for {
		select {
		case p, ok := <-progress:
			if !ok {
				break LOOP
			}
			switch p.State {
			case nosql.EvolutionStateError:
				break LOOP
			case nosql.EvolutionStateCompleted:
				break LOOP
			case nosql.EvolutionStatePreparing:
				fmt.Println("preparing...")
			case nosql.EvolutionStatePrepared:
				fmt.Println("prepared")
			case nosql.EvolutionStateDroppingIndex:
				fmt.Printf("dropping indexes: %.1f%s", p.IndexDrops.Pct()*100, "%\n")
				fmt.Printf("total progress: %.1f%s", p.Pct()*100, "%\n")
			case nosql.EvolutionStateCreatingIndex:
				fmt.Printf("creating indexes: %.1f%s", p.IndexCreates.Pct()*100, "%\n")
				fmt.Printf("total progress: %.1f%s", p.Pct()*100, "%\n")
			case nosql.EvolutionStateDroppingCollection:
				fmt.Printf("dropping collections: %.1f%s", p.CollectionDrops.Pct()*100, "%\n")
				fmt.Printf("total progress: %.1f%s", p.Pct()*100, "%\n")
			}
		}
	}
}

type MySchema struct {
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
		Num  nosql.Int64  `@:"num"`
		Key  nosql.String `@:"key"`
		Name nosql.String `@:"name"`
		//Names struct {
		//	First nosql.String `sort:"ASC"`
		//	Last  nosql.String `sort:"DESC"`
		//} `@:"[name.first,age,children.0]"` // Composite index
	}
}

type Schema2 struct {
	*nosql.Schema
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
