package nosql

import (
	"encoding/json"
	"fmt"
	"github.com/moontrade/mdbx-go"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	StoreName = "nosql"
	Version   = "0.1.0"
)

const (
	schemaCollectionID  = CollectionID(1)
	minUserCollectionID = CollectionID(100)
)

type schemasStore struct {
	store           *Store
	schemas         []*SchemaMeta
	schemasByUID    map[string]*SchemaMeta
	maxCollectionId uint32
	maxIndexID      uint32
	maxSchemaID     uint32
	mu              sync.Mutex
}

func (m *schemasStore) findMaxSchemaID() uint32 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.findMaxSchemaID0()
}

func (m *schemasStore) findMaxSchemaID0() uint32 {
	if len(m.schemas) == 0 {
		return 0
	}
	if len(m.schemas) == 1 {
		m.maxSchemaID = m.schemas[0].Id
		return m.maxSchemaID
	}
	max := uint32(0)
	for _, schema := range m.schemas {
		if schema.Id > max {
			max = schema.Id
		}
	}
	atomic.StoreUint32(&m.maxSchemaID, max)
	return atomic.LoadUint32(&m.maxSchemaID)
}

func (m *schemasStore) nextSchemaID() uint32 {
	return atomic.AddUint32(&m.maxSchemaID, 1)
}

func (m *schemasStore) findMaxIndexID() uint32 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.findMaxSchemaID0()
}

func (m *schemasStore) findMaxIndexID0() uint32 {
	if len(m.schemas) == 0 {
		return 0
	}
	max := uint32(0)
	for _, schema := range m.schemas {
		if len(schema.Collections) == 0 {
			continue
		}
		for _, col := range schema.Collections {
			if len(col.Indexes) == 0 {
				continue
			}
			for _, index := range col.Indexes {
				if index.ID > max {
					max = index.ID
				}
			}
		}
	}
	atomic.StoreUint32(&m.maxSchemaID, max)
	return m.maxSchemaID
}

func (m *schemasStore) nextIndexID() uint32 {
	return atomic.AddUint32(&m.maxIndexID, 1)
}

func (m *schemasStore) findMaxCollectionID() CollectionID {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.findMaxCollectionID0()
}

func (m *schemasStore) findMaxCollectionID0() CollectionID {
	max := CollectionID(0)
	for _, schema := range m.schemas {
		for _, col := range schema.Collections {
			if col.Id > max {
				max = col.Id
			}
		}
	}
	atomic.StoreUint32(&m.maxCollectionId, uint32(max))
	return CollectionID(atomic.LoadUint32(&m.maxCollectionId))
}

func (m *schemasStore) nextCollectionID() CollectionID {
	m.mu.Lock()
	defer m.mu.Unlock()

	max := CollectionID(m.maxCollectionId)
	if max == 0 {
		max = m.findMaxCollectionID0()
	}
	if max == 0 {
		atomic.StoreUint32(&m.maxCollectionId, uint32(minUserCollectionID))
		return CollectionID(atomic.LoadUint32(&m.maxCollectionId))
	}
	// Fast path
	if max < 32768 {
		return CollectionID(atomic.AddUint32(&m.maxCollectionId, 1))
	}

	// Create a sorted list of all IDs.
	ids := make([]uint16, 0, 128)
	for _, schema := range m.schemas {
		for _, col := range schema.Collections {
			ids = append(ids, uint16(col.Id))
		}
	}
	if len(ids) == 0 {
		m.maxCollectionId = 1
		return 1
	}
	sort.Sort(uint16Slice(ids))

	// Take the first gap
	next := uint16(1)
	for _, id := range ids {
		if id != next {
			return CollectionID(next)
		}
		next++
	}
	// Out of IDs?
	if atomic.LoadUint32(&m.maxCollectionId) == math.MaxUint16 {
		return 0
	}
	// Add a new ID
	return CollectionID(atomic.AddUint32(&m.maxCollectionId, 1))
}

func loadMeta(s *Store) (*schemasStore, error) {
	m := &schemasStore{
		store:        s,
		schemasByUID: make(map[string]*SchemaMeta),
	}

	// Use an update
	if err := s.store.View(func(tx *mdbx.Tx) error {
		var (
			cursor *mdbx.Cursor
			err    error
		)
		defer func() {
			if cursor != nil {
				_ = cursor.Close()
			}
		}()

		// ParseSchema Schema Records
		{
			var (
				k    = NewRecordID(schemaCollectionID, 0)
				key  = docIDVal(&k)
				data = mdbx.Val{}
			)

			cursor, err = tx.OpenCursor(s.documentsDBI)
			if err != mdbx.ErrSuccess {
				return err
			}

			// First record is reserved to describe the type of database this MDBX file is
			if err = cursor.Get(&key, &data, mdbx.CursorFirst); err != mdbx.ErrSuccess {
				if err == mdbx.ErrNotFound {
					// no schemas exist
					return nil
				}
			}

			addSchema := func() bool {
				k = DocID(key.UInt64())
				if k.CollectionID() != schemaCollectionID {
					return false
				}
				schema := &SchemaMeta{}
				if err = json.Unmarshal(data.Unsafe(), schema); err != nil {
					return false
				}

				m.schemas = append(m.schemas, schema)
				m.schemasByUID[schema.UID] = schema
				return true
			}

			if !addSchema() {
				return err
			}

		loop:
			for {
				if err = cursor.Get(&key, &data, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
					if err == mdbx.ErrNotFound {
						break loop
					}
				}
				if !addSchema() {
					break loop
				}
			}

			_ = cursor.Close()
			cursor = nil

			if err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	m.findMaxSchemaID()
	m.findMaxCollectionID()
	m.findMaxIndexID()
	return m, nil
}

// ChangeSet is a set of ChangeActions required to perform in order to get the
// schema consistent.
type ChangeSet struct {
	Started       time.Time
	From          *SchemaMeta
	To            *SchemaMeta
	Creates       []*CollectionCreate
	Drops         []*CollectionDrop
	IndexCreates  []*IndexCreate
	IndexRebuilds []*IndexCreate
	IndexDrops    []*IndexDrop
	mu            sync.Mutex
}

func (cs *ChangeSet) Apply(progress chan float32) error {
	return nil
}

type ChangeKind int

const (
	ChangeCreate       ChangeKind = 1
	ChangeDrop         ChangeKind = 2
	ChangeCreateIndex  ChangeKind = 3
	ChangeRebuildIndex ChangeKind = 4
	ChangeDropIndex    ChangeKind = 5
)

type ChangeAction interface {
	apply(batchSize int, progress chan float32) error
}

type changeAction struct{}

type CollectionCreate struct {
	meta  CollectionMeta
	store *collectionStore
}

// CollectionDrop action to delete all of a collection's documents and remove
// the metadata from the schema.
type CollectionDrop struct {
	meta  CollectionMeta
	store *collectionStore
}

type IndexCreate struct {
	meta  IndexMeta
	store *indexStore
}

type IndexRebuild struct {
	from  IndexMeta
	to    IndexMeta
	store *indexStore
}

// IndexDrop represents an action to drop an index from a schema, delete
// all data in the index database and remove the metadata from the saved schema.
type IndexDrop struct {
	meta  IndexMeta
	store *indexStore
}

func (m *schemasStore) load(nextSchema *Schema) (*ChangeSet, error) {
	cs := &ChangeSet{
		To: nextSchema.buildMeta(),
	}
	m.mu.Lock()
	cs.From = m.schemasByUID[nextSchema.Meta.UID]
	m.mu.Unlock()

	if cs.From != nil {
		existingCollections := make(map[string]CollectionMeta)
		for _, col := range cs.From.Collections {
			existingCollections[col.Name] = col
		}

		nextCollections := make(map[string]*collectionStore)
		collections := make([]*collectionStore, len(nextSchema.Collections))
		for i, col := range nextSchema.Collections {
			collections[i] = col.collectionStore
			if collections[i] == nil {
				return nil, ErrCollectionStore
			}
			if nextCollections[col.Name] != nil {
				return nil, fmt.Errorf("duplicate collection name used: %s", col.Name)
			}
			nextCollections[col.Name] = col.collectionStore
		}

		for _, col := range collections {
			existingCollection, ok := existingCollections[col.Name]
			if !ok {
				cs.Creates = append(cs.Creates, &CollectionCreate{
					meta:  col.CollectionMeta,
					store: col,
				})
			} else {
				// Was anything changed on collection?
				if !col.CollectionMeta.Equals(&existingCollection.collectionDescriptor) {
					// NOOP
				}

				existingIndexes := make(map[string]IndexMeta)
				for _, index := range existingCollection.Indexes {
					existingIndexes[index.Name] = index
				}
				var indexCreates []*IndexCreate
				var indexRebuilds []*IndexRebuild
				for _, index := range col.indexes {
					to := index.Meta()
					from, ok := existingIndexes[index.Name()]
					if !ok {
						indexCreates = append(indexCreates, &IndexCreate{
							meta:  index.Meta(),
							store: index.getStore(),
						})
					} else {
						delete(existingIndexes, index.Name())
						to.ID = from.ID
						if to.didChange(from) {
							indexRebuilds = append(indexRebuilds, &IndexRebuild{
								from:  from,
								to:    to,
								store: index.getStore(),
							})
						}
					}
				}

				var indexDrops []*IndexDrop
				for _, index := range existingIndexes {
					indexDrops = append(indexDrops, &IndexDrop{
						meta:  index,
						store: nil,
					})
				}

				if len(indexCreates) > 0 {
					cs.IndexCreates = append(cs.IndexCreates, indexCreates...)
				}
				if len(indexDrops) > 0 {
					cs.IndexDrops = append(cs.IndexDrops, indexDrops...)
				}

				// check for Index creates
				// check for Index drops

				// Remove from existingCollections map.
				delete(existingCollections, col.Name)
			}
		}

		// Anything remaining in existingCollections map needs to be dropped.
		if len(existingCollections) > 0 {
			for _, collection := range existingCollections {
				cs.Drops = append(cs.Drops, &CollectionDrop{
					meta:  collection,
					store: nil,
				})

				// Drop all indexes
				if len(collection.Indexes) > 0 {
					for _, index := range collection.Indexes {
						cs.IndexDrops = append(cs.IndexDrops, &IndexDrop{
							meta:  index,
							store: nil,
						})
					}
				}
			}
		}
	} else {
		cs.Creates = make([]*CollectionCreate, len(nextSchema.Collections))
		cs.IndexCreates = make([]*IndexCreate, 0, len(nextSchema.Collections))

		for i, col := range nextSchema.Collections {
			if col.collectionStore == nil {
				col.collectionStore = &collectionStore{
					store: m.store,
				}
			}
			col.collectionStore.store = m.store
			cs.Creates[i] = &CollectionCreate{
				meta:  col.CollectionMeta,
				store: col.collectionStore,
			}

			for _, index := range col.indexes {
				index.getStore().store = m.store
				cs.IndexCreates = append(cs.IndexCreates, &IndexCreate{
					meta:  index.Meta(),
					store: index.getStore(),
				})
			}
		}
	}

	return cs, nil
}
