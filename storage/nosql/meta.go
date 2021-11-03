package nosql

import (
	"github.com/moontrade/mdbx-go"
)

const (
	StoreName = "nosql"
	Version   = "0.1.0"

	mainSchema = ""
)

type storeRecord struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

const (
	metaRecordKindCollection = 1
	metaRecordKindIndex      = 2
)

type meta struct {
	store       *Store
	collections []*collectionStore
}

type schemaMeta struct {
	Name     string `json:"name"`
	Pkg      string `json:"pkg"`
	FQN      string `json:"fqn"`
	Checksum uint64 `json:"checksum"`
}

type collectionMeta struct {
	Id       uint16         `json:"id"`
	Created  uint64         `json:"c"`
	Updated  uint64         `json:"u"`
	Schema   int32          `json:"s"`
	Kind     CollectionKind `json:"k"`
	Name     string         `json:"n"`
	Indexes  []int32        `json:"ix"`
	Checksum uint64         `json:"x"`
}

func loadMeta(s *Store) (*meta, error) {
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
