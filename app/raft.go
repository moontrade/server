package app

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"github.com/moontrade/server/logger"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var errLeaderUnknown = errors.New("leader unknown")

type raftWrap struct {
	*raft.Raft
	conf      Config
	advertise string
	mu        sync.RWMutex
	extra     map[string]serverExtra
}

func (ra *raftWrap) getExtraForAddr(addr string) (extra serverExtra, ok bool) {
	if ra.advertise == "" {
		return extra, false
	}
	ra.mu.RLock()
	defer ra.mu.RUnlock()
	for eaddr, extra := range ra.extra {
		if eaddr == addr || extra.advertise == addr ||
			extra.remoteAddr == addr {
			return extra, true
		}
	}
	return extra, false
}

func raftInit(conf Config, hclogger hclog.Logger, fsm raft.FSM,
	logStore raft.LogStore, stableStore raft.StableStore,
	snaps raft.SnapshotStore, trans raft.Transport,
) *raftWrap {
	rconf := raft.DefaultConfig()
	rconf = &raft.Config{
		ProtocolVersion:    raft.ProtocolVersionMax,
		HeartbeatTimeout:   2000 * time.Millisecond,
		ElectionTimeout:    2000 * time.Millisecond,
		CommitTimeout:      200 * time.Millisecond,
		MaxAppendEntries:   1024,
		ShutdownOnRemove:   true,
		TrailingLogs:       10240,
		SnapshotInterval:   120 * time.Second,
		SnapshotThreshold:  8192,
		LeaderLeaseTimeout: 1000 * time.Millisecond,
		LogLevel:           "WARN",
	}
	rconf.Logger = hclogger
	rconf.LocalID = raft.ServerID(conf.NodeID)
	ra, err := raft.NewRaft(rconf, fsm, logStore, stableStore, snaps, trans)
	if err != nil {
		logger.Fatal(err)
	}
	return &raftWrap{
		Raft:      ra,
		conf:      conf,
		advertise: conf.Advertise,
	}
}

func getLeaderAdvertiseAddr(ra *raftWrap) string {
	leader := string(ra.Leader())
	if ra.advertise == "" {
		return leader
	}
	if leader == "" {
		return ""
	}
	extra, ok := ra.getExtraForAddr(leader)
	if !ok {
		return ""
	}
	return extra.advertise
}

func errRaftConvert(ra *raftWrap, err error) error {
	if ra.conf.TryErrors {
		if err == raft.ErrNotLeader {
			leader := getLeaderAdvertiseAddr(ra)
			if leader != "" {
				return fmt.Errorf("TRY %s", leader)
			}
		}
		return err
	}
	switch err {
	case raft.ErrNotLeader, raft.ErrLeadershipLost,
		raft.ErrLeadershipTransferInProgress:
		leader := getLeaderAdvertiseAddr(ra)
		if leader != "" {
			return fmt.Errorf("MOVED 0 %s", leader)
		}
		fallthrough
	case raft.ErrRaftShutdown, raft.ErrTransportShutdown:
		return fmt.Errorf("CLUSTERDOWN %s", err)
	}
	return err
}

// joinClusterIfNeeded attempts to make this server join a Raft cluster. If
// the server already belongs to a cluster or if the server is bootstrapping
// then this operation is ignored.
func joinClusterIfNeeded(conf Config, ra *raftWrap, addr net.Addr, tlscfg *tls.Config) {
	// Get the current Raft cluster configuration for determining whether this
	// server needs to bootstrap a new cluster, or join/re-join an existing
	// cluster.
	f := ra.GetConfiguration()
	if err := f.Error(); err != nil {
		log.Fatalf("could not get Raft configuration: %v", err)
	}
	var addrStr string
	if ra.advertise != "" {
		addrStr = conf.Advertise
	} else {
		addrStr = addr.String()
	}
	cfg := f.Configuration()
	servers := cfg.Servers
	if len(servers) == 0 {
		// Empty configuration. Either bootstrap or join an existing cluster.
		if conf.JoinAddr == "" {
			// No '-join' flag provided.
			// Bootstrap new cluster.
			logger.Notice("bootstrapping new cluster")

			var configuration raft.Configuration
			configuration.Servers = []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(conf.NodeID),
					Address:  raft.ServerAddress(addrStr),
				},
			}
			err := ra.BootstrapCluster(configuration).Error()
			if err != nil && err != raft.ErrCantBootstrap {
				log.Fatalf("bootstrap: %s", err)
			}
		} else {
			// Joining an existing cluster
			joinAddr := conf.JoinAddr
			logger.Notice("joining existing cluster at %v", joinAddr)
			err := func() error {
				for {
					conn, err := RedisDial(joinAddr, conf.Auth, tlscfg)
					if err != nil {
						return err
					}
					defer conn.Close()
					res, err := redis.String(conn.Do("raft", "server", "add",
						conf.NodeID, addrStr))
					if err != nil {
						if strings.HasPrefix(err.Error(), "MOVED ") {
							parts := strings.Split(err.Error(), " ")
							if len(parts) == 3 {
								joinAddr = parts[2]
								time.Sleep(time.Millisecond * 100)
								continue
							}
						}
						return err
					}
					if res != "1" {
						return fmt.Errorf("'1', got '%s'", res)
					}
					return nil
				}
			}()
			if err != nil {
				log.Fatalf("raft server add: %v", err)
			}
		}
	} else {
		if conf.JoinAddr != "" {
			logger.Warn("ignoring join request because server already " +
				"belongs to a cluster")
		}
		if ra.advertise != "" {
			// Check that the address is the same as before
			found := false
			same := true
			before := ra.advertise
			for _, s := range servers {
				if string(s.ID) == conf.NodeID {
					found = true
					if string(s.Address) != ra.advertise {
						same = false
						before = string(s.Address)
						break
					}
				}
			}
			if !found {
				log.Fatalf("advertise address changed but node not found\n")
			} else if !same {
				log.Fatalf("advertise address change from \"%s\" to \"%s\" ",
					before, ra.advertise)
			}
		}
	}
}

func getClusterLastIndex(ra *raftWrap, tlscfg *tls.Config, auth string,
) (uint64, error) {
	if ra.State() == raft.Leader {
		return ra.LastIndex(), nil
	}
	addr := getLeaderAdvertiseAddr(ra)
	if addr == "" {
		return 0, errLeaderUnknown
	}
	conn, err := RedisDial(addr, auth, tlscfg)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	args, err := redis.Strings(conn.Do("raft", "info", "last_log_index"))
	if err != nil {
		return 0, err
	}
	if len(args) != 2 {
		return 0, errors.New("invalid response")
	}
	lastIndex, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return lastIndex, nil
}
