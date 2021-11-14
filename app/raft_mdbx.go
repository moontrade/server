package app

import (
	"errors"
	"github.com/moontrade/server/logger"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/hashicorp/raft"
	"github.com/moontrade/mdbx-go"
)

var (
	ErrPathNotDir = errors.New("path is not a directory")
)

const (
	DefaultLogFlags = mdbx.EnvNoMetaSync |
		mdbx.EnvNoTLS |
		mdbx.EnvWriteMap |
		mdbx.EnvLIFOReclaim |
		mdbx.EnvNoMemInit |
		mdbx.EnvCoalesce

	DefaultStableFlags = mdbx.EnvSyncDurable |
		mdbx.EnvNoTLS |
		mdbx.EnvWriteMap |
		mdbx.EnvLIFOReclaim |
		mdbx.EnvNoMemInit |
		mdbx.EnvCoalesce

	Kilobyte = 1024
	Megabyte = 1024 * 1024
	Gigabyte = Megabyte * 1024
	Terabyte = Gigabyte * 1024
)

var (
	_ raft.LogStore    = (*Store)(nil)
	_ raft.StableStore = (*Store)(nil)

	DefaultStableGeometry = mdbx.Geometry{
		SizeLower:       64 * Kilobyte,
		SizeNow:         8 * Kilobyte,
		SizeUpper:       256 * Kilobyte,
		GrowthStep:      64 * Kilobyte,
		ShrinkThreshold: 128 * Kilobyte,
		PageSize:        4 * Kilobyte,
	}
	DefaultLogGeometry = mdbx.Geometry{
		SizeLower:       1 * Megabyte,
		SizeNow:         1 * Megabyte,
		SizeUpper:       4 * Gigabyte,
		GrowthStep:      16 * Megabyte,
		ShrinkThreshold: 8 * Megabyte,
		PageSize:        8 * Kilobyte,
	}
)

const (
	raftStableDBI   = "raftstable"
	raftLogDBI      = "raftlog"
	keyCurrentTerm  = "CurrentTerm"
	keyLastVoteTerm = "LastVoteTerm"
	keyLastVoteCand = "LastVoteCand"
)

type Store struct {
	logStore     *mdbx.Store
	stableStore  *mdbx.Store
	logDBI       mdbx.DBI
	stableDBI    mdbx.DBI
	stableCache  map[string][]byte
	stableUint64 map[string]uint64
	firstIndex   uint64
	lastIndex    uint64
	currentTerm  uint64
	lastVoteTerm uint64
	lastVoteCand []byte
	mu           sync.RWMutex
}

func OpenStore(path string, logFlags, stableFlags mdbx.EnvFlags, mode os.FileMode) (*Store, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err = os.MkdirAll(path, mode); err != nil {
			return nil, err
		}
	} else if !stat.IsDir() {
		return nil, ErrPathNotDir
	}

	s := &Store{
		stableCache:  make(map[string][]byte),
		stableUint64: make(map[string]uint64),
	}

	if s.logStore, err = mdbx.Open(filepath.Join(path, "log"), logFlags, mode,
		func(env *mdbx.Env, create bool) error {
			//if create {
			if e := env.SetMaxDBS(1); e != mdbx.ErrSuccess {
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
				if s.logDBI, e = tx.OpenDBI(raftLogDBI, mdbx.DBCreate|mdbx.DBIntegerKey); e != mdbx.ErrSuccess {
					return e
				}
				return nil
			})
		}); err != nil {
		return nil, err
	}

	if s.stableStore, err = mdbx.Open(filepath.Join(path, "stable"), stableFlags, mode,
		func(env *mdbx.Env, create bool) error {
			if e := env.SetMaxDBS(1); e != mdbx.ErrSuccess {
				return e
			}
			// Set geometry
			if e := env.SetGeometry(DefaultStableGeometry); e != mdbx.ErrSuccess {
				return e
			}
			return nil
		}, func(store *mdbx.Store, create bool) error {
			return store.Update(func(tx *mdbx.Tx) error {
				var e mdbx.Error
				if s.stableDBI, e = tx.OpenDBI(raftStableDBI, mdbx.DBCreate); e != mdbx.ErrSuccess {
					return e
				}
				return nil
			})
		}); err != nil {
		_ = s.logStore.Close()
		return nil, err
	}

	// Load stable store
	if e := s.loadStableStore(); e != nil && e != mdbx.ErrSuccess {
		_ = s.Close()
		return nil, err
	}
	// Load first and last index
	if e := s.loadFirstAndLastIndex(); e != nil && e != mdbx.ErrSuccess {
		_ = s.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	err := s.logStore.Close()
	if s.stableStore != s.logStore {
		e := s.stableStore.Close()
		if err != nil {
			if e != nil {
				return err
			}
			return err
		} else if e != nil {
			return e
		}
	}
	return err
}

func (s *Store) loadFirstAndLastIndex() error {
	return s.logStore.View(func(tx *mdbx.Tx) error {
		var (
			cursor, err = tx.OpenCursor(s.logDBI)
			k, v        mdbx.Val
		)
		if err != mdbx.ErrSuccess {
			return err
		}
		defer cursor.Close()

		// Find FirstIndex
		if err = cursor.Get(&k, &v, mdbx.CursorFirst); err != mdbx.ErrSuccess {
			if err == mdbx.ErrNotFound {
				return nil
			}
			return err
		}
		if k.Base != nil && k.Len == 8 {
			atomic.StoreUint64(&s.firstIndex, *(*uint64)(unsafe.Pointer(k.Base)))
		}

		// Find LastIndex
		if err = cursor.Get(&k, &v, mdbx.CursorLast); err != mdbx.ErrSuccess {
			if err == mdbx.ErrNotFound {
				return nil
			}
			return err
		}
		if k.Base != nil && k.Len == 8 {
			atomic.StoreUint64(&s.lastIndex, *(*uint64)(unsafe.Pointer(k.Base)))
		}
		return nil
	})
}

func (s *Store) loadStableStore() error {
	return s.stableStore.View(func(tx *mdbx.Tx) error {
		var (
			key  = mdbx.StringConst(keyCurrentTerm)
			data = mdbx.Val{}
		)

		if e := tx.Get(s.stableDBI, &key, &data); e != mdbx.ErrSuccess {
			if e != mdbx.ErrNotFound {
				return e
			}
		} else if data.Base != nil && data.Len == 8 {
			atomic.StoreUint64(&s.currentTerm, *(*uint64)(unsafe.Pointer(data.Base)))
		}

		key = mdbx.StringConst(keyLastVoteTerm)
		if err := tx.Get(s.stableDBI, &key, &data); err != mdbx.ErrSuccess {
			if err != mdbx.ErrNotFound {
				return err
			}
		} else if data.Base != nil && data.Len == 8 {
			atomic.StoreUint64(&s.lastVoteTerm, *(*uint64)(unsafe.Pointer(data.Base)))
		}

		key = mdbx.StringConst(keyLastVoteCand)
		if err := tx.Get(s.stableDBI, &key, &data); err != mdbx.ErrSuccess {
			if err != mdbx.ErrNotFound {
				return err
			}
		} else if data.Base != nil && data.Len > 0 {
			s.lastVoteCand = make([]byte, data.Len)
			copy(s.lastVoteCand, data.UnsafeBytes())
		}
		return nil
	})
}

func (s *Store) Set(key []byte, val []byte) error {
	if e := s.stableStore.Update(func(tx *mdbx.Tx) error {
		var (
			k = mdbx.Bytes(&key)
			v = mdbx.Bytes(&val)
		)
		return tx.Put(s.stableDBI, &k, &v, 0)
	}); e != nil && e != mdbx.ErrSuccess {
		return e
	}

	ks := *(*string)(unsafe.Pointer(&key))
	switch ks {
	case keyLastVoteCand:
		s.lastVoteCand = val
	default:
		s.mu.Lock()
		s.stableCache[ks] = val
		s.mu.Unlock()
	}
	return nil
}

// Get returns the value for key, or an empty byte slice if key was not found.
func (s *Store) Get(key []byte) (result []byte, err error) {
	// Fast cache
	ks := *(*string)(unsafe.Pointer(&key))
	switch ks {
	case keyLastVoteCand:
		return s.lastVoteCand, nil
	}

	s.mu.RLock()
	v, ok := s.stableCache[ks]
	s.mu.RUnlock()
	if ok {
		return v, nil
	}
	if e := s.stableStore.View(func(tx *mdbx.Tx) error {
		var (
			k = mdbx.Bytes(&key)
			v = mdbx.Val{}
		)
		e := tx.Get(s.stableDBI, &k, &v)
		if e != mdbx.ErrSuccess {
			return e
		}
		result = make([]byte, v.Len)
		copy(result, v.UnsafeBytes())
		return nil
	}); e != nil {
		if e == mdbx.ErrNotFound {
			return nil, nil
		}
		return nil, e
	}

	s.mu.Lock()
	s.stableCache[ks] = result
	s.mu.Unlock()
	return
}

func (s *Store) SetUint64(key []byte, val uint64) error {
	if e := s.stableStore.Update(func(tx *mdbx.Tx) error {
		var (
			k = mdbx.Bytes(&key)
			v = mdbx.Val{
				Base: (*byte)(unsafe.Pointer(&val)),
				Len:  8,
			}
		)
		return tx.Put(s.stableDBI, &k, &v, 0)
	}); e != nil && e != mdbx.ErrSuccess {
		return e
	}

	ks := *(*string)(unsafe.Pointer(&key))
	switch ks {
	case keyCurrentTerm:
		atomic.StoreUint64(&s.currentTerm, val)
	case keyLastVoteTerm:
		atomic.StoreUint64(&s.lastVoteTerm, val)
	default:
		s.mu.Lock()
		s.stableUint64[ks] = val
		s.mu.Unlock()
	}
	return nil
}

// GetUint64 returns the uint64 value for key, or 0 if key was not found.
func (s *Store) GetUint64(key []byte) (result uint64, err error) {
	// Fast cache
	ks := *(*string)(unsafe.Pointer(&key))
	switch ks {
	case keyCurrentTerm:
		return atomic.LoadUint64(&s.currentTerm), nil
	case keyLastVoteTerm:
		return atomic.LoadUint64(&s.lastVoteTerm), nil
	}

	// Cached in map
	s.mu.RLock()
	cached, ok := s.stableUint64[ks]
	s.mu.RUnlock()
	if ok {
		return cached, nil
	}

	if e := s.stableStore.View(func(tx *mdbx.Tx) error {
		var (
			k   = mdbx.Bytes(&key)
			val = mdbx.Val{}
		)
		e := tx.Get(s.stableDBI, &k, &val)
		if e != mdbx.ErrSuccess {
			return e
		}
		if val.Base != nil && val.Len >= 8 {
			result = *(*uint64)(unsafe.Pointer(val.Base))
		} else {
			result = 0
		}
		return nil
	}); e != nil && e != mdbx.ErrSuccess {
		if e == mdbx.ErrNotFound {
			return 0, nil
		}
		return 0, e
	}

	// Set cache
	s.mu.Lock()
	s.stableUint64[ks] = result
	s.mu.Unlock()
	return
}

// FirstIndex returns the first index written. 0 for no entries.
func (s *Store) FirstIndex() (uint64, error) {
	return atomic.LoadUint64(&s.firstIndex), nil
}

// LastIndex returns the last index written. 0 for no entries.
func (s *Store) LastIndex() (uint64, error) {
	return atomic.LoadUint64(&s.lastIndex), nil
}

// GetLog gets a log entry at a given index.
func (s *Store) GetLog(index uint64, log *raft.Log) error {
	if err := s.logStore.View(func(tx *mdbx.Tx) error {
		var (
			key = mdbx.Val{
				Base: (*byte)(unsafe.Pointer(&index)),
				Len:  8,
			}
			val = mdbx.Val{}
		)

		if e := tx.Get(s.logDBI, &key, &val); e != mdbx.ErrSuccess {
			return e
		}

		if val.Base != nil {
			if err := unmarshalLog(val.UnsafeBytes(), log); err != nil {
				return err
			}
		}

		return nil
	}); err != nil && err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return raft.ErrLogNotFound
		}
		return err
	}
	return nil
}

// StoreLog stores a log entry.
func (s *Store) StoreLog(log *raft.Log) error {
	if err := s.logStore.Update(func(tx *mdbx.Tx) error {
		var (
			k = mdbx.Val{
				Base: (*byte)(unsafe.Pointer(&log.Index)),
				Len:  8,
			}
			v = mdbx.Val{
				Len: uint64(SerializedSize(log)),
			}
		)

		// PutReserve to save a Go allocation.
		if e := tx.Put(s.logDBI, &k, &v, mdbx.PutReserve|mdbx.PutAppend); e != mdbx.ErrSuccess {
			return e
		}
		// Marshal directly into MDBX managed buffer which will be saved once committed.
		if _, err := marshalLog(log, v.UnsafeBytes()); err != nil {
			return err
		}
		return nil
	}); err != nil && err != mdbx.ErrSuccess {
		return err
	}

	if log.Index == 1 {
		atomic.StoreUint64(&s.firstIndex, log.Index)
	}
	atomic.StoreUint64(&s.lastIndex, log.Index)
	return nil
}

// StoreLogs stores multiple log entries.
func (s *Store) StoreLogs(logs []*raft.Log) error {
	if len(logs) == 0 {
		return nil
	}
	if err := s.logStore.Update(func(tx *mdbx.Tx) error {
		cursor, e := tx.OpenCursor(s.logDBI)
		if e != mdbx.ErrSuccess {
			return e
		}
		defer cursor.Close()
		var err error

		for _, log := range logs {
			var (
				k = mdbx.Val{
					Base: (*byte)(unsafe.Pointer(&log.Index)),
					Len:  8,
				}
				v = mdbx.Val{
					Len: uint64(SerializedSize(log)),
				}
			)

			// PutReserve to save a Go allocation.
			if err = cursor.Put(&k, &v, mdbx.PutReserve|mdbx.PutAppend); e != mdbx.ErrSuccess {
				return err
			}
			if _, err = marshalLog(log, v.UnsafeBytes()); err != nil {
				return err
			}
		}

		return nil
	}); err != nil && err != mdbx.ErrSuccess {
		return err
	}

	atomic.StoreUint64(&s.lastIndex, logs[len(logs)-1].Index)

	// sync to disk
	if err := s.logStore.Sync(); err != nil {
		logger.WarnErr(err)
	}
	return nil
}

// DeleteRange deletes a range of log entries. The range is inclusive.
func (s *Store) DeleteRange(min, max uint64) error {
	if err := s.logStore.Update(func(tx *mdbx.Tx) error {
		cursor, err := tx.OpenCursor(s.logDBI)
		if err != mdbx.ErrSuccess {
			return err
		}
		defer cursor.Close()

		for index := min; index <= max; index++ {
			var (
				k = mdbx.Val{
					Base: (*byte)(unsafe.Pointer(&index)),
					Len:  8,
				}
				v = mdbx.Val{}
			)

			// PutReserve to save a Go allocation.
			if err = cursor.Get(&k, &v, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
				if err == mdbx.ErrNotFound {
					// TODO: Is this correct?
					return nil
				}
				return err
			}
			if err = cursor.Delete(0); err != mdbx.ErrSuccess {
				return err
			}
		}

		return nil
	}); err != nil && err != mdbx.ErrSuccess {
		return err
	}

	atomic.StoreUint64(&s.firstIndex, max+1)
	return nil
}

func (s *Store) Sync() error {
	err := s.stableStore.Env().Sync(true, false)
	err2 := s.logStore.Env().Sync(true, false)
	if err != mdbx.ErrSuccess {
		return err
	}
	if err2 != mdbx.ErrSuccess {
		return err2
	}
	return nil
}

func (s *Store) SyncStable() error {
	if err := s.stableStore.Env().Sync(true, true); err != mdbx.ErrSuccess {
		return err
	}
	return nil
}

func (s *Store) SyncLog() error {
	if err := s.stableStore.Env().Sync(true, false); err != mdbx.ErrSuccess {
		return err
	}
	return nil
}

func SerializedSize(log *raft.Log) int {
	if log == nil {
		return 0
	}
	return 33 + len(log.Data) + len(log.Extensions)
}

func marshalLog(log *raft.Log, b []byte) ([]byte, error) {
	sz := SerializedSize(log)
	if len(b) < sz {
		b = make([]byte, sz)
	} else {
		b = b[0:sz]
	}
	*(*uint64)(unsafe.Pointer(&b[0])) = log.Index
	*(*uint64)(unsafe.Pointer(&b[8])) = log.Term
	if log.AppendedAt.IsZero() {
		*(*int64)(unsafe.Pointer(&b[16])) = 0
	} else {
		appendedAt := log.AppendedAt.UTC().UnixNano()
		*(*int64)(unsafe.Pointer(&b[16])) = appendedAt
	}
	b[24] = byte(log.Type)
	*(*uint32)(unsafe.Pointer(&b[25])) = uint32(len(log.Data))
	*(*uint32)(unsafe.Pointer(&b[29])) = uint32(len(log.Extensions))
	if len(log.Data) > 0 {
		copy(b[33:], log.Data)
	}
	if len(log.Extensions) > 0 {
		copy(b[33+len(log.Data):], log.Extensions)
	}
	return b, nil
}

func unmarshalLog(b []byte, log *raft.Log) error {
	if len(b) < 33 {
		return errors.New("malformed log message")
	}
	log.Index = *(*uint64)(unsafe.Pointer(&b[0]))
	log.Term = *(*uint64)(unsafe.Pointer(&b[8]))
	appendedAt := *(*int64)(unsafe.Pointer(&b[16]))
	if appendedAt > 0 {
		log.AppendedAt = time.Unix(appendedAt/1000000000, appendedAt%1000000000)
	}
	log.Type = raft.LogType(b[24])
	dataLength := *(*uint32)(unsafe.Pointer(&b[25]))
	extensionLength := *(*uint32)(unsafe.Pointer(&b[29]))
	total := dataLength + extensionLength
	if total == 0 {
		return nil
	}
	if len(b) < int(total)+33 {
		return io.ErrShortBuffer
	}
	if dataLength > 0 {
		if len(log.Data) >= int(dataLength) {
			log.Data = log.Data[0:dataLength]
		} else {
			log.Data = make([]byte, dataLength)
		}
		copy(log.Data, b[33:33+dataLength])
	} else {
		if log.Data != nil {
			log.Data = log.Data[:0]
		}
	}

	if extensionLength > 0 {
		if len(log.Extensions) >= int(extensionLength) {
			log.Extensions = log.Extensions[0:extensionLength]
		} else {
			log.Extensions = make([]byte, extensionLength)
		}
		copy(log.Extensions, b[33+dataLength:33+dataLength+extensionLength])
	} else {
		if log.Extensions != nil {
			log.Extensions = log.Extensions[:0]
		}
	}

	return nil
}
