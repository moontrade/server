package app

import (
	"compress/gzip"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"github.com/moontrade/server/logger"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"time"
)

func snapshotInit(conf Config, dir string, m *machine, hclogger hclog.Logger) raft.SnapshotStore {
	snaps, err := raft.NewFileSnapshotStoreWithLogger(dir, 3, hclogger)
	if err != nil {
		logger.Fatal(err)
	}
	m.snaps = snaps
	return snaps
}

// A Snapshot is an interface that allows for Raft snapshots to be taken.
type Snapshot interface {
	Persist(io.Writer) error
	Done(path string)
}

type fsmSnap struct {
	id    string
	dir   string
	snap  Snapshot
	ts    int64
	seed  int64
	start int64
}

func (s *fsmSnap) Persist(sink raft.SnapshotSink) error {
	s.id = sink.ID()
	gw := gzip.NewWriter(sink)
	var head [32]byte
	copy(head[:], "SNAP0001")
	binary.LittleEndian.PutUint64(head[8:], uint64(s.start))
	binary.LittleEndian.PutUint64(head[16:], uint64(s.ts))
	binary.LittleEndian.PutUint64(head[24:], uint64(s.seed))
	n, err := gw.Write(head[:])
	if err != nil {
		return err
	}
	if n != 32 {
		return errors.New("invalid write")
	}
	if err := s.snap.Persist(gw); err != nil {
		return err
	}
	return gw.Close()
}

func (s *fsmSnap) Release() {
	path := filepath.Join(s.dir, "snapshots", s.id, "state.bin")
	if _, err := readSnapInfo(s.id, path); err != nil {
		path = ""
	}
	s.snap.Done(path)
}

func (m *machine) Snapshot() (raft.FSMSnapshot, error) {
	snapshot := m.snapshot
	if snapshot == nil {
		if m.jsonSnaps {
			snapshot = jsonSnapshot
		} else {
			return nil, errors.New("snapshots are disabled")
		}
	}
	usnap, err := snapshot(m.Data())
	if err != nil {
		return nil, err
	}
	snap := &fsmSnap{
		dir:   m.dir,
		snap:  usnap,
		seed:  m.seed,
		ts:    m.ts,
		start: m.start,
	}
	return snap, nil
}

func readSnapHead(r io.Reader) (start, ts, seed int64, err error) {
	var head [32]byte
	n, err := io.ReadFull(r, head[:])
	if err != nil {
		return 0, 0, 0, err
	}
	if n != 32 {
		return 0, 0, 0, errors.New("invalid read")
	}
	if string(head[:8]) != "SNAP0001" {
		return 0, 0, 0, errors.New("invalid snapshot signature")
	}
	start = int64(binary.LittleEndian.Uint64(head[8:]))
	ts = int64(binary.LittleEndian.Uint64(head[16:]))
	seed = int64(binary.LittleEndian.Uint64(head[24:]))
	return start, ts, seed, nil
}

func (m *machine) Restore(rc io.ReadCloser) error {
	restore := m.restore
	if restore == nil {
		if m.jsonSnaps {
			restore = func(rd io.Reader) (data interface{}, err error) {
				return jsonRestore(rd, m.jsonType)
			}
		} else {
			return errors.New("snapshot restoring is disabled")
		}
	}
	gr, err := gzip.NewReader(rc)
	if err != nil {
		return err
	}
	start, ts, seed, err := readSnapHead(gr)
	if err != nil {
		return err
	}
	m.start = start
	m.ts = ts
	m.seed = seed
	m.data, err = restore(gr)
	return err
}

type restoreData struct {
	data  interface{}
	ts    int64
	seed  int64
	start int64
}

func dataDirInit(conf Config) (string, *restoreData) {
	var rdata *restoreData
	dir := filepath.Join(conf.DataDir, conf.Name, conf.NodeID)
	if conf.BackupPath != "" {
		_, err := os.Stat(dir)
		if err == nil {
			logger.Warn("backup restore ignored: "+
				"data directory already exists: path=%s", dir)
			return dir, nil
		}
		logger.Print("restoring backup: path=%s", conf.BackupPath)
		if !os.IsNotExist(err) {
			logger.Fatal(err)
		}
		rdata, err = dataDirRestoreBackup(conf, dir)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Print("recovery successful")
	} else {
		if err := os.MkdirAll(dir, 0777); err != nil {
			logger.Fatal(err)
		}
	}
	if conf.DataDirReady != nil {
		conf.DataDirReady(dir)
	}
	return dir, rdata
}

func dataDirRestoreBackup(conf Config, dir string) (rdata *restoreData, err error) {
	rdata = new(restoreData)
	f, err := os.Open(conf.BackupPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	rdata.start, rdata.ts, rdata.seed, err = readSnapHead(gr)
	if err != nil {
		return nil, err
	}
	if conf.Restore != nil {
		rdata.data, err = conf.Restore(gr)
		if err != nil {
			return nil, err
		}
	} else if conf.jsonSnaps {
		rdata.data, err = func(rd io.Reader) (data interface{}, err error) {
			return jsonRestore(rd, conf.jsonType)
		}(gr)
		if err != nil {
			return nil, err
		}
	} else {
		rdata.data = conf.InitialData
	}
	return rdata, nil
}

type jsonSnapshotType struct{ jsdata []byte }

func (s *jsonSnapshotType) Done(path string) {}
func (s *jsonSnapshotType) Persist(wr io.Writer) error {
	_, err := wr.Write(s.jsdata)
	return err
}
func jsonSnapshot(data interface{}) (Snapshot, error) {
	if data == nil {
		return &jsonSnapshotType{}, nil
	}
	jsdata, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &jsonSnapshotType{jsdata: jsdata}, nil
}

func jsonRestore(rd io.Reader, typ reflect.Type) (interface{}, error) {
	jsdata, err := ioutil.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	if typ == nil {
		return nil, nil
	}
	data := reflect.New(typ).Interface()
	if err = json.Unmarshal(jsdata, data); err != nil {
		return nil, err
	}
	return data, err
}

// runLogLoadedPoller is a background routine that reports on raft log progress
// and also maintains the m.logLoaded atomic boolean for open read systems.
func runLogLoadedPoller(conf Config, m *machine, ra *raftWrap, tlscfg *tls.Config) {
	var loaded bool
	var lastPerc string
	lastPrint := time.Now()
	for {
		// load the last index from the cluster leader
		lastIndex, err := getClusterLastIndex(ra, tlscfg, conf.Auth)
		if err != nil {
			if err != errLeaderUnknown {
				logger.Warn("cluster_last_index: %v", err)
			} else {
				// This service is probably a candidate, flip the loaded
				// off to begin printing log progress.
				loaded = false
				atomic.StoreInt32(&m.logLoaded, 0)
			}
			time.Sleep(time.Second)
			continue
		}

		// update machine with the known leader last index and determine
		// the load progress and how many logs are remaining.
		m.mu.Lock()
		m.logRemain = lastIndex - m.appliedIndex
		if lastIndex == 0 {
			m.logPercent = 0
		} else {
			m.logPercent = float64(m.appliedIndex-m.firstIndex) /
				float64(lastIndex-m.firstIndex)
		}
		lpercent := m.logPercent
		remain := m.logRemain
		m.mu.Unlock()

		if !loaded {
			// Print progress status to console log
			perc := fmt.Sprintf("%.1f%%", lpercent*100)
			if remain < 5 {
				logger.Print("logs loaded: ready for commands")
				loaded = true
				atomic.StoreInt32(&m.logLoaded, 1)
			} else if perc != "0.0%" && perc != lastPerc {
				msg := fmt.Sprintf("logs progress: %.1f%%, remaining=%d",
					lpercent*100, remain)
				now := time.Now()
				if now.Sub(lastPrint) > time.Second*5 {
					logger.Print(msg)
					lastPrint = now
				} else {
					logger.Info(msg)
				}
			}
			lastPerc = perc
		}
		time.Sleep(time.Second / 5)
	}
}
