package nosql

import (
	"github.com/moontrade/mdbx-go"
)

const (
	StoreName = "nosql"
	Version   = "0.1.0"
)

type storeRecord struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

const (
	metaRecordKindJsonCollection      = 1
	metaRecordKindJsonCollectionIndex = 2
)

const (
	dbiCollectionU64 = 1
	dbiIndexU64      = 2
	dbiIndex         = 2
)

type metaRecord struct {
	name        string   `json:"n,omitempty"`
	version     string   `json:"v,omitempty"`
	description string   `json:"d,omitempty"`
	schema      string   `json:"s,omitempty"`
	kind        int      `json:"k,omitempty"`
	dbi         int      `json:"dbi"`
	children    []uint16 `json:"c,omitempty"`
	parent      uint16   `json:"p,omitempty"`
	size        uint32   `json:"sz,omitempty"`
	fixedSize   uint32   `json:"fsz,omitempty"`
	count       uint32   `json:"cnt,omitempty"`
}

type meta struct {
	store       *Store
	collections []*collectionStore
}

func loadMeta(s *Store, schema *Schema) (*meta, error) {
	m := &meta{store: s}

	//schemaCollections := make(map[string]*Collection)
	//
	//if schema != nil {
	//	for _, c := range schema.Collections {
	//		if c == nil {
	//			return nil, errors.New("schema: nil collection")
	//		}
	//		c.Name = strings.ToLower(c.Name)
	//		if _, ok := schemaCollections[c.Name]; ok {
	//			return nil, errors.Errorf("schema: duplicate collection named %s", c.Name)
	//		}
	//		schemaCollections[c.Name] = c
	//	}
	//}

	//existingJson := make(map[string]*collectionStore)
	//existingJson := make(map[string]*collectionStore)

	// Use an update
	if err := s.store.View(func(tx *mdbx.Tx) error {
		cursor, err := tx.OpenCursor(s.metaDBI)
		if err != mdbx.ErrSuccess {
			return err
		}
		defer func() {
			if cursor != nil {
				_ = cursor.Close()
			}
		}()

		var (
			key  = mdbx.Val{}
			data = mdbx.Val{}
		)

		// First record is reserved to describe the type of database this MDBX file is
		if err = cursor.Get(&key, &data, mdbx.CursorFirst); err != mdbx.ErrSuccess {
			if err == mdbx.ErrNotFound {
				// Insert storeRecord
			}
		}

	loop:
		for {
			if err = cursor.Get(&key, &data, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
				if err == mdbx.ErrNotFound {
					break loop
				}
			}

		}

		return nil
	}); err != nil {
		return nil, err
	}
	return m, nil
}
