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
		schema   = &Schema{UID: "@"}
		schema2  = &Schema2{UID: "@"}
		progress <-chan nosql.EvolutionProgress
	)

	if progress, err = store.HydrateTyped(context.Background(), schema); err != nil {
		t.Fatal(err)
	}
	if err = wait(progress); err != nil {
		t.Fatal(err)
	}
	if progress, err = store.HydrateTyped(context.Background(), schema2); err != nil {
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
		schema   = &Schema{}
		progress <-chan nosql.EvolutionProgress
	)
	if progress, err = store.HydrateTyped(context.Background(), schema); err != nil {
		t.Fatal(err)
	}
	if err = wait(progress); err != nil {
		t.Fatal(err)
	}

	if err = schema.Update(func(tx *nosql.Tx) error {
		if err := schema.Orders.Insert(tx, &Order{
			Num:   100,
			Key:   "ORD1",
			Price: 1.7843,
		}); err != nil {
			return err
		}
		if err := schema.Orders.Insert(tx, &Order{
			Num:   101,
			Key:   "ORD2",
			Price: 1.8912,
		}); err != nil {
			return err
		}
		if err := schema.Orders.Insert(tx, &Order{
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

type Schema struct {
	*nosql.Schema
	UID    string
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
	UID    string
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
	ID    nosql.DocID `json:"_id"`
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

func (orders *Orders) Update(tx *nosql.Tx, order *Order) error {
	return orders.Collection.Update(tx, order.ID, order, nil, nil)
}

func (orders *Orders) Delete(tx *nosql.Tx, id nosql.DocID, order *Order) (bool, error) {
	if id == 0 {
		id = order.ID
	}
	return orders.Collection.Delete(tx, id, order, nil)
}
