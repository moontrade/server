package app

import (
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/moontrade/server/logger"
)

func storeInit(conf Config, dir string) (raft.LogStore, raft.StableStore) {
	//conf.Backend = MDBX
	switch conf.Backend {
	//case Bolt:
	//	store, err := raftboltdb.NewBoltStore(filepath.Join(dir, "store.db"))
	//	if err != nil {
	//		log.Fatalf("bolt store open: %s", err)
	//	}
	//	return store, store
	//
	//case LevelDB:
	//	dur := raftleveldb.High
	//	if conf.NoSync {
	//		dur = raftleveldb.Medium
	//	}
	//	store, err := raftleveldb.NewLevelDBStore(
	//		filepath.Join(dir, "store"), dur)
	//	if err != nil {
	//		log.Fatalf("leveldb store open: %s", err)
	//	}
	//	return store, store

	case MDBX:
		store, err := OpenStore(
			dir,
			DefaultLogFlags,
			DefaultStableFlags,
			0755,
		)
		if err != nil {
			logger.Fatal(fmt.Errorf("mdbx store open: %s", err))
		}
		return store, store

	default:
		logger.Fatal(errors.New("invalid backend"))
	}
	return nil, nil
}
