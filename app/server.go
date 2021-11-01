package app

import (
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"github.com/moontrade/server/logger"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type serverExtra struct {
	reachable  bool   // server is reachable
	remoteAddr string // remote tcp address
	advertise  string // advertise address
	lastError  error  // last error, if any
}

type serverEntry struct {
	id      string
	address string
	resolve string
	leader  bool
}

func (e *serverEntry) clusterID() string {
	src := sha1.Sum([]byte(e.id))
	return hex.EncodeToString(src[:])
}

func (e *serverEntry) host() string {
	idx := strings.LastIndexByte(e.address, ':')
	if idx == -1 {
		return ""
	}
	return e.address[:idx]
}

func (e *serverEntry) port() int {
	idx := strings.LastIndexByte(e.address, ':')
	if idx == -1 {
		return 0
	}
	port, _ := strconv.Atoi(e.address[idx+1:])
	return port
}

func (ra *raftWrap) getServerList() ([]serverEntry, error) {
	leader := string(ra.Leader())
	f := ra.GetConfiguration()
	err := f.Error()
	if err != nil {
		return nil, err
	}
	cfg := f.Configuration()
	var servers []serverEntry
	for _, s := range cfg.Servers {
		var entry serverEntry
		entry.id = string(s.ID)
		entry.address = string(s.Address)
		extra, ok := ra.getExtraForAddr(entry.address)
		if ok {
			entry.resolve = extra.remoteAddr
		} else {
			entry.resolve = entry.address
		}
		entry.leader = entry.resolve == leader || entry.address == leader
		servers = append(servers, entry)
	}
	return servers, nil
}

func runMaintainServers(ra *raftWrap) {
	if ra.advertise == "" {
		return
	}
	for {
		f := ra.GetConfiguration()
		if err := f.Error(); err != nil {
			time.Sleep(time.Second)
			continue
		}
		cfg := f.Configuration()
		var wg sync.WaitGroup
		wg.Add(len(cfg.Servers))
		for _, svr := range cfg.Servers {
			go func(addr string) {
				defer wg.Done()
				c, err := net.DialTimeout("tcp", addr, time.Second*5)
				if err == nil {
					defer c.Close()
				}
				ra.mu.Lock()
				defer ra.mu.Unlock()
				if ra.extra == nil {
					ra.extra = make(map[string]serverExtra)
				}
				extra := ra.extra[addr]
				if err != nil {
					extra.reachable = false
					extra.lastError = err
				} else {
					extra.reachable = true
					extra.lastError = nil
					extra.remoteAddr = c.RemoteAddr().String()
					extra.advertise = addr
				}
				ra.extra[addr] = extra
			}(string(svr.Address))
		}
		wg.Wait()
		time.Sleep(time.Second)
	}
}

func serverInit(conf Config, tlscfg *tls.Config) (*splitServer, net.Addr) {
	var ln net.Listener
	var err error
	if tlscfg != nil {
		ln, err = tls.Listen("tcp4", conf.Addr, tlscfg)
	} else {
		ln, err = net.Listen("tcp4", conf.Addr)
	}
	if err != nil {
		logger.Fatal(err)
	}
	logger.Print("server listening at %s", ln.Addr())
	if conf.Advertise != "" {
		logger.Print("server advertising as %s", conf.Advertise)
	}
	if conf.ServerReady != nil {
		conf.ServerReady(ln.Addr().String(), conf.Auth, tlscfg)
	}
	return newSplitServer(ln), ln.Addr()
}

func parseTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	pair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	tlscfg := &tls.Config{
		Certificates: []tls.Certificate{pair},
	}
	for _, cert := range pair.Certificate {
		pcert, err := x509.ParseCertificate(cert)
		if err != nil {
			return nil, err
		}
		if len(pcert.DNSNames) > 0 {
			tlscfg.ServerName = pcert.DNSNames[0]
			break
		}
	}
	return tlscfg, nil
}

func tlsInit(conf Config) *tls.Config {
	if conf.TLSCertPath == "" || conf.TLSKeyPath == "" {
		return nil
	}
	tlscfg, err := parseTLSConfig(conf.TLSCertPath, conf.TLSKeyPath)
	if err != nil {
		logger.Fatal(err)
	}
	return tlscfg
}

// splitServer split a single server socket/listener into multiple logical
// listeners. For our use case, there is one transport listener and one client
// listener sharing the same server socket.
type splitServer struct {
	ln net.Listener
	//log      zerolog.Logger
	matchers []*matcher
}

func newSplitServer(ln net.Listener) *splitServer {
	return &splitServer{ln: ln}
}

func (m *splitServer) serve() error {
	for {
		c, err := m.ln.Accept()
		if err != nil {
			logger.Error(err)
			continue
		}
		conn := &conn{Conn: c, matching: true}
		var matched bool
		for _, ma := range m.matchers {
			conn.bufpos = 0
			if n, ok := ma.sniff(conn); ok {
				conn.buffer = conn.buffer[n:]
				conn.matching = false
				ma.ln.next <- conn
				matched = true
				break
			}
		}
		if !matched {
			c.Close()
		}
	}
}

func (m *splitServer) split(sniff func(r io.Reader) (n int, ok bool),
) net.Listener {
	ln := &listener{addr: m.ln.Addr(), next: make(chan net.Conn)}
	m.matchers = append(m.matchers, &matcher{sniff, ln})
	return ln
}

type matcher struct {
	sniff func(r io.Reader) (n int, matched bool)
	ln    *listener
}

type conn struct {
	net.Conn
	matching bool
	buffer   []byte
	bufpos   int
}

func (c *conn) Read(p []byte) (n int, err error) {
	if c.matching {
		// matching mode
		if c.bufpos == len(c.buffer) {
			// need more buffer
			packet := make([]byte, 4096)
			nn, err := c.Conn.Read(packet)
			if err != nil {
				return 0, err
			}
			if nn == 0 {
				return 0, nil
			}
			c.buffer = append(c.buffer, packet[:nn]...)
		}
		copy(p, c.buffer[c.bufpos:])
		if len(p) < len(c.buffer)-c.bufpos {
			n = len(p)
		} else {
			n = len(c.buffer) - c.bufpos
		}
		c.bufpos += n
		return n, nil
	}
	if len(c.buffer) > 0 {
		// normal mode but with a buffer
		copy(p, c.buffer)
		if len(p) < len(c.buffer) {
			n = len(p)
			c.buffer = c.buffer[len(p):]
			if len(c.buffer) == 0 {
				c.buffer = nil
			}
		} else {
			n = len(c.buffer)
			c.buffer = nil
		}
		return n, nil
	}
	// normal mode, no buffer
	return c.Conn.Read(p)
}

// listener is a split network listener
type listener struct {
	addr net.Addr
	next chan net.Conn
}

func (l *listener) Accept() (net.Conn, error) {
	return <-l.next, nil
}
func (l *listener) Addr() net.Addr {
	return l.addr
}
func (l *listener) Close() error {
	return errors.New("disabled")
}
