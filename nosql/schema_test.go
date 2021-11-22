package nosql_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/moontrade/server/nosql"
	"math"
	"testing"
)

func TestDocID(t *testing.T) {
	var (
		id1 = nosql.NewDocID(100, 2)
		id2 = nosql.NewDocID(100, 1)
		id3 = nosql.NewDocID(1, 1)
	)

	fmt.Println(id1)
	fmt.Println(id2)
	fmt.Println(id3)

	fmt.Println(uint64(math.MaxUint64) / uint64(math.MaxUint16))
}

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
	if err = wait(progress); err != nil {
		t.Fatal(err)
	}

	if schema, err = nosql.ParseSchemaWithUID("@", mySchema2); err != nil {
		t.Fatal(err)
	}
	if progress, err = store.Hydrate(context.Background(), schema); err != nil {
		t.Fatal(err)
	}
	if err = wait(progress); err != nil {
		t.Fatal(err)
	}
}

func TestCRUD(t *testing.T) {
	store, err := nosql.Open(&nosql.Config{
		Path:  "./testdata",
		Flags: nosql.DefaultDurable,
		Mode:  0755,
	})
	if err != nil {
		t.Fatal(err)
	}

	var (
		mySchema = &MySchema{}
		schema   *nosql.Schema
		progress <-chan nosql.EvolutionProgress
	)
	if schema, err = nosql.ParseSchemaWithUID("@", mySchema); err != nil {
		t.Fatal(err)
	}
	if progress, err = store.Hydrate(context.Background(), schema); err != nil {
		t.Fatal(err)
	}
	if err = wait(progress); err != nil {
		t.Fatal(err)
	}

	if err = mySchema.Update(func(tx *nosql.Tx) error {
		if err := mySchema.Orders.Insert(tx, &Order{
			Num:   100,
			Key:   "ORD1",
			Price: 1.7843,
		}); err != nil {
			return err
		}
		if err := mySchema.Orders.Insert(tx, &Order{
			Num:   101,
			Key:   "ORD2",
			Price: 1.8912,
		}); err != nil {
			return err
		}
		if err := mySchema.Orders.Insert(tx, &Order{
			Num:   102,
			Key:   "ORD3",
			Price: 1.9758,
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func wait(progress <-chan nosql.EvolutionProgress) error {
	if progress == nil {
		return errors.New("schema does not require an evolution")
	}
LOOP:
	for {
		select {
		case p, ok := <-progress:
			if !ok {
				return p.Err
			}
			switch p.State {
			case nosql.EvolutionStateError:
				return p.Err
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
	return nil
}

type MySchema struct {
	*nosql.Schema

	Orders Orders

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
	ID    nosql.DocID `json:"ID"`
	Num   uint64      `json:"num"`
	Key   string      `json:"key"`
	Price float64     `json:"price"`
}

type Orders struct {
	_ Order
	nosql.Collection
	Num       nosql.Int64Unique  `@:"num"`
	Key       nosql.StringUnique `@:"key"`
	Price     nosql.Float64      `@:"price"`
	FirstName nosql.String       `@:"name.first"`
}

func (orders *Orders) Insert(tx *nosql.Tx, order *Order) error {
	order.ID = orders.NextID()
	return orders.Collection.Insert(tx, order.ID, order, nil)
}
