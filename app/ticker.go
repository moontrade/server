package app

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/hashicorp/raft"
	"github.com/moontrade/server/logger"
	"strconv"
	"time"
)

// runTicker is a background routine that keeps the raft machine time and
// random seed updated.
func runTicker(conf Config, rt *remoteTime, m *machine, ra *raftWrap) {
	rbuf := make([]byte, 4096)
	var rnb []byte
	for {
		start := time.Now()
		ts := rt.Now().UnixNano()
		if len(rnb) == 0 {
			n, err := rand.Read(rbuf[:])
			if err != nil || n != len(rbuf) {
				logger.Panic(err)
			}
			rnb = rbuf[:]
		}
		seed := int64(binary.LittleEndian.Uint64(rnb))
		rnb = rnb[8:]
		req := new(writeRequestFuture)
		req.args = []string{
			"tick",
			strconv.FormatInt(ts, 10),
			strconv.FormatInt(seed, 10),
		}
		req.wg.Add(1)
		m.wrC <- req
		req.wg.Wait()
		m.mu.Lock()
		if req.err == nil {
			l := req.resp.(raft.Log)
			m.tickedIndex = l.Index
			m.tickedTerm = l.Term
		} else {
			m.tickedIndex = 0
			m.tickedTerm = 0
		}
		m.tickedSig.Broadcast()
		m.mu.Unlock()
		dur := time.Since(start)
		delay := m.tickDelay - dur
		if delay < 1 {
			delay = 1
		}
		time.Sleep(delay)
	}
}
