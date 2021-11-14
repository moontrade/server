# Embedded Replicated NoSQL Database

NoSQL is a concise and simple document oriented database layer on top of libmdbx BTree MMAP storage engine.

## Schemas

### Collections

Collections are just buckets of document blobs identified by a monotonic 64bit key. Layout below:

```
DocumentID (64-bits)
    CollectionID 16-bits
    RowID        48-bits
```

Given CollectionID is 16-bits and the first 100 IDs are reserved for system uses, that equates to 65436 being the max number of collections.

The format of a document is not enforced. However, in order to support secondary indexes selectors that create document projections are required.

### Indexes

Collections may have any number of secondary indexes. Of course, the more indexes you create the slower writes will become and the higher the storage requirements. Below is a list of index data types:

- Int64
- Float64
- String
- Full-Text (not implemented)
- Geo (not implemented)

### Streams

Streams are a specialized type of collection that supports Kafka like streams and time-series like streams.


### Evolution

Schema evolutions are a set of actions that must be performed on the database in order for the schema to be in sync. Below are a list of action types:

- Drop Collection
  - Delete all documents
- Create Collection
  - Assign 16-bit CollectionID
- Create Index
  - Assign 32-bit IndexID
  - Iterate through all Collection documents and build index records
- Rebuild Index
  - Delete all index records
  - Iterate through all Collection documents and build index records
- Drop Index
  - Delete all index records



## Schema Mapping

Schemas may be mapped using a Schema prototype expressed in code.

```go
import "github.com/moontrade/server/nosql"

var DB = &Schema{}

type Schema struct {
	*nosql.Schema // auto-generated nosql.Schema instance during Hydrate
	
	Orders struct {
		_ Order // optional document struct
		nosql.Collection // required
		Num       nosql.Int64Unique  `@:"num"`
		Key       nosql.StringUnique `@:"key"`
		Price     nosql.Float64      `@:"price"`
	}
	
	Contact struct {
		FirstName nosql.String      `@:"name.first"`
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

func main() {
	
}
```


## libmdbx Storage Layout


