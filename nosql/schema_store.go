package nosql

import (
	"context"
	"encoding/json"
	"github.com/moontrade/mdbx-go"
	"math"
	"sort"
	"sync"
	"sync/atomic"
)

const (
	StoreName = "nosql"
	Version   = "0.1.0"
)

const (
	schemaCollectionID  = CollectionID(1)
	minUserCollectionID = CollectionID(100)
)

// schemasStore manages all schemas in a Store.
type schemasStore struct {
	store           *Store
	schemas         []*SchemaMeta
	schemasByUID    map[string]*SchemaMeta
	maxCollectionId uint32
	maxIndexID      uint32
	maxSchemaID     uint32
	mu              sync.Mutex
}

func (s *Store) Hydrate(ctx context.Context, schema *Schema) (chan EvolutionProgress, error) {
	//if schema.IsLoaded() {
	//	return nil, ErrAlreadyLoaded
	//}
	return s.schemas.hydrate(ctx, schema)
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
	return atomic.LoadUint32(&m.maxSchemaID)
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

// nextCollectionID finds the next collection ID. It first will try incrementing
// the latest Max if under a threshold. Above that threshold, it creates a sorted
// list of all collection IDs then scans for the first monotonic gap.
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

func openSchemaStore(s *Store) (*schemasStore, error) {
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
				k    = NewDocID(schemaCollectionID, 0)
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
				k = DocID(key.U64())
				if k.CollectionID() != schemaCollectionID {
					return false
				}
				schema := &SchemaMeta{}
				if err = json.Unmarshal(data.UnsafeBytes(), schema); err != nil {
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
	}); err != nil && err != mdbx.ErrSuccess {
		return nil, err
	}

	m.findMaxSchemaID()
	m.findMaxCollectionID()
	m.findMaxIndexID()
	return m, nil
}
