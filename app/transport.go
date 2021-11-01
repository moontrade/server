package app

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"github.com/moontrade/server/logger"
	"io"
	"net"
	"time"
)

// RedisDial is a helper function that dials out to another Uhaha server with
// redis protocol and using the provded TLS config and Auth token. The TLS/Auth
// must be correct in order to establish a connection.
func RedisDial(addr, auth string, tlscfg *tls.Config) (redis.Conn, error) {
	var conn redis.Conn
	var err error
	if tlscfg != nil {
		conn, err = redis.Dial("tcp", addr,
			redis.DialUseTLS(true), redis.DialTLSConfig(tlscfg))
	} else {
		conn, err = redis.Dial("tcp", addr)
	}
	if err != nil {
		return nil, err
	}
	if auth != "" {
		res, err := redis.String(conn.Do("auth", auth))
		if err != nil {
			conn.Close()
			return nil, err
		}
		if res != "OK" {
			conn.Close()
			return nil, fmt.Errorf("'OK', got '%s'", res)
		}
	}
	return conn, nil
}

const transportMarker = "8e35747e37d192d9a819021ba2a02909"

type transportStream struct {
	net.Listener
	auth   string
	tlscfg *tls.Config
}

func (s *transportStream) Dial(addr raft.ServerAddress, timeout time.Duration) (conn net.Conn, err error) {
	if timeout <= 0 {
		if s.tlscfg != nil {
			conn, err = tls.Dial("tcp", string(addr), s.tlscfg)
		} else {
			conn, err = net.Dial("tcp", string(addr))
		}
	} else {
		if s.tlscfg != nil {
			conn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeout},
				"tcp", string(addr), s.tlscfg)
		} else {
			conn, err = net.DialTimeout("tcp", string(addr), timeout)
		}
	}
	if err != nil {
		return nil, err
	}
	if _, err := conn.Write([]byte(transportMarker)); err != nil {
		conn.Close()
		return nil, err
	}
	if _, err := conn.Write([]byte(s.auth)); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func transportInit(conf Config, tlscfg *tls.Config, svr *splitServer, hclogger hclog.Logger) raft.Transport {
	ln := svr.split(func(r io.Reader) (n int, ok bool) {
		rd := bufio.NewReader(r)
		for i := 0; i < len(transportMarker); i++ {
			b, err := rd.ReadByte()
			if err != nil || b != transportMarker[i] {
				return 0, false
			}
		}
		for i := 0; i < len(conf.Auth); i++ {
			b, err := rd.ReadByte()
			if err != nil || b != conf.Auth[i] {
				return 0, false
			}
		}
		return len(transportMarker) + len(conf.Auth), true
	})
	stream := new(transportStream)
	stream.Listener = ln
	stream.auth = conf.Auth
	stream.tlscfg = tlscfg
	return raft.NewNetworkTransport(stream, conf.MaxPool, 0, logger.RaftWriter)
}
