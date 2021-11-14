package nosql

import (
	"errors"
	"github.com/moontrade/mdbx-go"
	"os"
	"sync"
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
	DefaultDurable = mdbx.EnvSyncDurable |
		mdbx.EnvNoTLS |
		mdbx.EnvLIFOReclaim |
		mdbx.EnvNoMemInit |
		mdbx.EnvCoalesce

	DefaultDurableFast = mdbx.EnvNoMetaSync |
		mdbx.EnvNoTLS |
		mdbx.EnvLIFOReclaim |
		mdbx.EnvNoMemInit |
		mdbx.EnvCoalesce

	DefaultAsync = mdbx.EnvSafeNoSync |
		mdbx.EnvNoTLS |
		mdbx.EnvLIFOReclaim |
		mdbx.EnvNoMemInit |
		mdbx.EnvCoalesce

	DefaultAsyncMap = mdbx.EnvSafeNoSync |
		mdbx.EnvWriteMap |
		mdbx.EnvNoTLS |
		mdbx.EnvLIFOReclaim |
		mdbx.EnvNoMemInit |
		mdbx.EnvCoalesce

	DefaultNoSync = mdbx.EnvUtterlyNoSync |
		mdbx.EnvWriteMap |
		mdbx.EnvNoTLS |
		mdbx.EnvLIFOReclaim |
		mdbx.EnvNoMemInit |
		mdbx.EnvCoalesce
)

const (
	kvDBI        = "kv"
	documentsDBI = "docs"
	indexDBI     = "idx"
)

var (
	Default *Store
)

// Store is a simple embedded Raft replicated ACID noSQL database
// built on MDBX B+Tree storage.
//
type Store struct {
	config       *Config
	store        *mdbx.Store // mdbx store
	kvDBI        mdbx.DBI    // generic Key/Value database
	documentsDBI mdbx.DBI    // documents database
	indexDBI     mdbx.DBI    // indexes database
	streamDBI    mdbx.DBI    // streams database
	schemas      *schemasStore
	mu           sync.Mutex
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.store == nil {
		return nil
	}
	if err := s.store.Close(); err != nil {
		return err
	}
	s.store = nil
	return nil
}

type Config struct {
	Path  string
	Flags mdbx.EnvFlags
	Mode  os.FileMode
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
			if e := env.SetMaxDBS(4); e != mdbx.ErrSuccess {
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
				if s.kvDBI, e = tx.OpenDBI(kvDBI, mdbx.DBCreate); e != mdbx.ErrSuccess {
					return e
				}
				if s.documentsDBI, e = tx.OpenDBIEx(documentsDBI, mdbx.DBCreate|mdbx.DBIntegerKey, mdbx.CmpU64, nil); e != mdbx.ErrSuccess {
					return e
				}
				if s.indexDBI, e = tx.OpenDBIEx(indexDBI, mdbx.DBCreate, mdbx.CmpU32PrefixLexical, nil); e != mdbx.ErrSuccess {
					return e
				}
				return nil
			})
		}); err != nil && err != mdbx.ErrSuccess {
		return nil, err
	}

	if s.schemas, err = openSchemaStore(s); err != nil {
		_ = s.store.Close()
		return nil, err
	}

	return s, nil
}
