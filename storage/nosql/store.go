package nosql

import (
	"errors"
	"github.com/moontrade/mdbx-go"
	"os"
)

var (
	DefaultLogGeometry = mdbx.Geometry{
		SizeLower:       1024 * 1024,
		SizeNow:         1024 * 1024,
		SizeUpper:       1024 * 1024 * 1024 * 16,
		GrowthStep:      1024 * 1024 * 4,
		ShrinkThreshold: 1024 * 1024 * 4 * 2,
		PageSize:        4096,
	}
)

const (
	metaDBI        = "meta"
	kvDBI          = "kv"
	collectionsDBI = "col"
	indexDBI       = "idx"
	indexDupeDBI   = "idxdupe"
)

// Store is a simple embedded Raft replicated noSQL database.
type Store struct {
	config         *Config
	store          *mdbx.Store // mdbx store
	metaDBI        mdbx.DBI    // meta-data database describing schema
	kvDBI          mdbx.DBI    // generic Key/Value database
	collectionsDBI mdbx.DBI    // collections database
	indexDBI       mdbx.DBI    // any byte slice indexes for collections
	//indexDupDBI    mdbx.DBI    // any byte slice indexes for collections
	schema *schemaStore
}

type Config struct {
	Path   string
	Flags  mdbx.EnvFlags
	Mode   os.FileMode
	Schema *Schema
}

func Open(config *Config) (*Store, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	var (
		s = &Store{
			config: config,
		}
		err error
	)

	if s.store, err = mdbx.Open(config.Path, config.Flags, config.Mode,
		func(env *mdbx.Env, create bool) error {
			//if create {
			if e := env.SetMaxDBS(5); e != mdbx.ErrSuccess {
				return e
			}
			// Set geometry
			if e := env.SetGeometry(DefaultLogGeometry); e != mdbx.ErrSuccess {
				return e
			}
			//}
			return nil
		}, func(store *mdbx.Store, create bool) error {
			return store.Update(func(tx *mdbx.Tx) error {
				var e mdbx.Error
				if s.collectionsDBI, e = tx.OpenDBI(collectionsDBI, mdbx.DBCreate|mdbx.DBIntegerKey); e != mdbx.ErrSuccess {
					return e
				}
				if s.indexDBI, e = tx.OpenDBI(indexDBI, mdbx.DBCreate); e != mdbx.ErrSuccess {
					return e
				}
				//if s.indexDupDBI, e = tx.OpenDBI(indexDupeDBI, mdbx.DBCreate); e != mdbx.ErrSuccess {
				//	return e
				//}
				if s.kvDBI, e = tx.OpenDBI(kvDBI, mdbx.DBCreate); e != mdbx.ErrSuccess {
					return e
				}
				return nil
			})
		}); err != nil {
		return nil, err
	}

	return nil, nil
}
