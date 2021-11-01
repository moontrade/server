package app

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/snappy"
	"github.com/hashicorp/raft"
	"github.com/moontrade/server/logger"
	"io"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func machineInit(conf Config, dir string, rdata *restoreData) *machine {
	m := new(machine)
	m.dir = dir
	m.vers = versline(conf)
	m.tickedSig = sync.NewCond(&m.mu)
	m.created = time.Now().UnixNano()
	m.wrC = make(chan *writeRequestFuture, 1024)
	m.tickDelay = conf.TickDelay
	m.openReads = conf.OpenReads
	if rdata != nil {
		m.data = rdata.data
		m.start = rdata.start
		m.seed = rdata.seed
		m.ts = rdata.ts
	} else {
		m.data = conf.InitialData
	}

	m.connClosed = conf.ConnClosed
	m.connOpened = conf.ConnOpened
	m.snapshot = conf.Snapshot
	m.restore = conf.Restore
	m.jsonSnaps = conf.jsonSnaps
	m.jsonType = conf.jsonType
	m.tick = conf.Tick
	m.commands = map[string]command{
		"tick":    {'w', cmdTICK},
		"barrier": {'w', cmdBARRIER},
		"raft":    {'s', cmdRAFT},
		"cluster": {'s', cmdCLUSTER},
		"machine": {'r', cmdMACHINE},
		"version": {'s', cmdVERSION},
	}
	if conf.TryErrors {
		delete(m.commands, "cluster")
	}
	for k, v := range conf.cmds {
		if _, ok := m.commands[k]; !ok {
			m.commands[k] = v
		}
	}
	m.catchall = conf.catchall
	return m
}

// The Machine interface is passed to every command. It includes the user data
// and various utilities that should be used from Write, Read, and Intermediate
// commands.
//
// It's important to note that the Data(), Now(), and Rand() functions can be
// used safely for Write and Read commands, but are not available for
// Intermediate commands. The Context() is ONLY available for Intermediate
// commands.
//
// A call to Rand() and Now() from inside of a Read command will always return
// back the same last known value of it's respective type. While, from a Write
// command, you'll get freshly generated values. This is to ensure that
// the every single command ALWAYS generates the same series of data on every
// server.
type Machine interface {
	// Data is the original user data interface that was assigned at startup.
	// It's safe to alter the data in this interface while inside a Write
	// command, but it's only safe to read from this interface for Read
	// commands.
	// Returns nil for Intermediate Commands.
	Data() interface{}
	// Now generates a stable timestamp that is synced with internet time
	// and for Write commands is always monotonical increasing. It's made to
	// be a trusted source of time for performing operations on the user data.
	// Always use this function instead of the builtin time.Now().
	// Returns nil for Intermediate Commands.
	Now() time.Time
	// Rand is a random number generator that must be used instead of the
	// standard Go packages `crypto/rand` and `math/rand`. For Write commands
	// the values returned from this generator are crypto seeded, guaranteed
	// to be reproduced in exact order when the server restarts, and identical
	// across all machines in the cluster. The underlying implementation is
	// PCG. Check out http://www.pcg-random.org/ for more information.
	// Returns nil for Intermediate Commands.
	Rand() Rand
	// Context returns the connection context that was defined in from the
	// Config.ConnOpened callback. Only available for Intermediate commands.
	// Returns nil for Read and Write Commands.
	Context() interface{}
}

type machine struct {
	snapshot   func(data interface{}) (Snapshot, error)
	restore    func(rd io.Reader) (data interface{}, err error)
	connOpened func(addr string) (context interface{}, accept bool)
	connClosed func(context interface{}, addr string)
	jsonSnaps  bool               //
	jsonType   reflect.Type       //
	snaps      raft.SnapshotStore //
	dir        string             //
	vers       string             // version line
	tick       func(m Machine)    //
	created    int64              // machine instance created timestamp
	commands   map[string]command // command table
	catchall   command            // catchall command
	openReads  bool               // open reads on by default
	tickDelay  time.Duration      // ticker delay

	mu           sync.RWMutex // protect all things in group
	firstIndex   uint64       // first applied index
	appliedIndex uint64       // last applied index (stable state)
	readers      int32        // (atomic counter) number of current readers
	tickedIndex  uint64       // index of last tick
	tickedTerm   uint64       // term of last tick
	tickedSig    *sync.Cond   // signal when ticked
	logPercent   float64      // percentage of log loaded
	logRemain    uint64       // non-applied log entries
	logLoaded    int32        // (atomic bool) log is loaded, allow open reads
	snap         bool         // snapshot in progress
	start        int64        // !! PERSISTED !! first non-zero timestamp
	ts           int64        // !! PERSISTED !! current timestamp
	seed         int64        // !! PERSISTED !! current seed
	data         interface{}  // !! PERSISTED !! user data

	wrC chan *writeRequestFuture
}

var _ Machine = &machine{}

type applyResp struct {
	resp interface{}
	elap time.Duration
	err  error
}

func (m *machine) Context() interface{} {
	return nil
}

func (m *machine) Apply(l *raft.Log) interface{} {
	packet, err := snappy.Decode(nil, l.Data)
	if err != nil {
		logger.Panic(err)
	}
	m.mu.Lock()
	defer func() {
		m.appliedIndex = l.Index
		if m.firstIndex == 0 {
			m.firstIndex = m.appliedIndex
		}
		m.mu.Unlock()
	}()
	numReqs, n := binary.Uvarint(packet)
	if n <= 0 {
		logger.Panic(errors.New("invalid apply"))
	}
	packet = packet[n:]
	resps := make([]applyResp, numReqs)
	for i := 0; i < int(numReqs); i++ {
		numArgs, n := binary.Uvarint(packet)
		if n <= 0 {
			logger.Panic(errors.New("invalid apply"))
		}
		packet = packet[n:]
		args := make([]string, numArgs)
		for i := 0; i < len(args); i++ {
			argLen, n := binary.Uvarint(packet)
			if n <= 0 {
				logger.Panic(errors.New("invalid apply"))
			}
			packet = packet[n:]
			args[i] = string(packet[:argLen])
			packet = packet[argLen:]
		}
		if len(args) == 0 {
			resps[i] = applyResp{nil, 0, nil}
		} else {
			cmdName := strings.ToLower(string(args[0]))
			cmd := m.commands[cmdName]
			if cmd.kind != 'w' {
				logger.Panic(fmt.Errorf("invalid apply '%c', command: '%s'",
					cmd.kind, cmdName))
			}
			tick := cmdName == "tick"
			if m.start == 0 && !tick {
				// This is in fact the leader, but because the machine has yet
				// to receive a valid tick command, we'll treat this as if the
				// server *is not* the leader.
				resps[i] = applyResp{nil, 0, raft.ErrNotLeader}
			} else {
				start := time.Now()
				res, err := cmd.fn(m, nil, args)
				if tick {
					// return only the index and term
					res = raft.Log{Index: l.Index, Term: l.Term}
				}
				resps[i] = applyResp{res, time.Since(start), err}
			}
		}
	}
	return resps
}

func (m *machine) Data() interface{} {
	return m.data
}

func (m *machine) Rand() Rand {
	return m
}

func (m *machine) Uint32() uint32 {
	seed := rincr(rincr(m.seed)) // twice called intentionally
	x := rgen(seed)
	if atomic.LoadInt32(&m.readers) == 0 {
		m.seed = seed
	}
	return x
}

func (m *machine) Uint64() uint64 {
	return (uint64(m.Uint32()) << 32) | uint64(m.Uint32())
}

func (m *machine) Int() int {
	return int(m.Uint64() << 1 >> 1)
}

func (m *machine) Float64() float64 {
	return float64(m.Uint32()) / 4294967296.0
}

func (m *machine) Read(p []byte) (n int, err error) {
	seed := rincr(m.seed)
	for len(p) >= 4 {
		seed = rincr(seed)
		binary.LittleEndian.PutUint32(p, rgen(seed))
		p = p[4:]
	}
	if len(p) > 0 {
		var last [4]byte
		seed = rincr(seed)
		binary.LittleEndian.PutUint32(last[:], rgen(seed))
		for i := 0; i < len(p); i++ {
			p[i] = last[i]
		}
	}
	if atomic.LoadInt32(&m.readers) == 0 {
		m.seed = seed
	}
	return len(p), nil
}

func (m *machine) Now() time.Time {
	ts := m.ts
	if atomic.LoadInt32(&m.readers) == 0 {
		m.ts++
	}
	return time.Unix(0, ts).UTC()
}

// intermediateMachine wraps the machine in a connection context
type intermediateMachine struct {
	context interface{}
	m       *machine
}

var _ Machine = intermediateMachine{}

func (m intermediateMachine) Now() time.Time       { return time.Time{} }
func (m intermediateMachine) Context() interface{} { return m.context }
func (m intermediateMachine) Rand() Rand           { return nil }
func (m intermediateMachine) Data() interface{}    { return nil }

func getBaseMachine(m Machine) *machine {
	switch m := m.(type) {
	case intermediateMachine:
		return m.m
	case *machine:
		return m
	default:
		return nil
	}
}

// RawMachineInfo represents the raw components of the machine
type RawMachineInfo struct {
	TS   int64
	Seed int64
}

// ReadRawMachineInfo reads the raw machine components.
func ReadRawMachineInfo(m Machine, info *RawMachineInfo) {
	*info = RawMachineInfo{}
	if m := getBaseMachine(m); m != nil {
		info.TS = m.ts
		info.Seed = m.seed
	}
}

// WriteRawMachineInfo writes raw components to the machine. Use with care as
// this operation may destroy the consistency of your cluster.
func WriteRawMachineInfo(m Machine, info *RawMachineInfo) {
	if m := getBaseMachine(m); m != nil {
		m.ts = info.TS
		m.seed = info.Seed
	}
}
