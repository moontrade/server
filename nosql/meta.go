package nosql

import (
	"encoding/json"
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
	store        *Store
	schemas      []*SchemaMeta
	schemasByUID map[string]*SchemaMeta
	maxColID     CollectionID
	maxIndexID   uint32
	maxSchemaID  uint32
	mu           sync.Mutex
}

func (m *schemasStore) nextSchemaID() uint32 {
	return atomic.AddUint32(&m.maxSchemaID, 1)
}

func (m *schemasStore) nextIndexID() uint32 {
	return atomic.AddUint32(&m.maxIndexID, 1)
}

func (m *schemasStore) findMaxColID() CollectionID {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.findMaxColID0()
}

func (m *schemasStore) findMaxColID0() CollectionID {
	ids := make([]uint16, 0, 128)
	for _, schema := range m.schemas {
		for _, col := range schema.Collections {
			ids = append(ids, uint16(col.Id))
		}
	}
	if len(ids) == 0 {
		return 0
	}
	sort.Sort(uint16Slice(ids))
	m.maxColID = CollectionID(ids[len(ids)-1])
	return m.maxColID
}

func (m *schemasStore) nextCollectionID() CollectionID {
	m.mu.Lock()
	defer m.mu.Unlock()

	max := m.maxColID
	if max == 0 {
		max = m.findMaxColID0()
	}
	if max == 0 {
		m.maxColID = minUserCollectionID
		return m.maxColID
	}
	// Fast path
	if max < 32768 {
		m.maxColID++
		return m.maxColID
	}

	// Create a sorted list of all IDs.
	ids := make([]uint16, 0, 128)
	for _, schema := range m.schemas {
		for _, col := range schema.Collections {
			ids = append(ids, uint16(col.Id))
		}
	}
	if len(ids) == 0 {
		m.maxColID = 1
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
	if m.maxColID == CollectionID(uint16(math.MaxUint16)) {
		return 0
	}
	// Add a new ID
	m.maxColID++
	return m.maxColID
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

	m.findMaxColID()
	return m, nil
}

type ChangeSet struct {
	Started       time.Time
	From          SchemaMeta
	To            SchemaMeta
	Creates       []*CollectionCreate
	Drops         []*CollectionDrop
	IndexCreates  []*IndexCreate
	IndexRebuilds []*IndexCreate
	IndexDrops    []*IndexCreate
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

func (m *schemasStore) load(parsed *Schema) (*ChangeSet, error) {
	m.mu.Lock()
	existing := m.schemasByUID[parsed.Meta.UID]
	m.mu.Unlock()

	cs := &ChangeSet{}

	if existing != nil {
		existingCollections := make(map[string]CollectionMeta)
		for _, col := range existing.Collections {
			existingCollections[col.Name] = col
		}
		collections := make([]*collectionStore, len(parsed.Collections))
		for i, col := range parsed.Collections {
			collections[i] = col.collectionStore
			if collections[i] == nil {
				return nil, ErrCollectionStore
			}
		}
		//for _, col := range collections {
		//	existingCollection, ok := existingCollections[col.Name]
		//	if !ok {
		//
		//	} else {
		//		delete(existingCollections, col.Name)
		//	}
		//}

		// Deletes
		if len(existingCollections) > 0 {

		}
	} else {
		//creates := make([]*collectionStore, 0, len(parsed.Collections))
		//drops := make([]*collectionStore, 0, len(parsed.Collections))
		//createIndexes := make([]*indexStore, 0, len(parsed.Collections))
		//dropIndexes := make([]*indexStore, 0, len(parsed.Collections))
	}

	return cs, nil
}
