package app

import (
	"errors"
	"github.com/hashicorp/raft"
	"github.com/moontrade/server/logger"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func startUserServices(conf Config, svr *splitServer, m *machine, ra *raftWrap) {
	// rearrange so that services with nil sniffers are last
	var nilServices []serviceEntry
	var services []serviceEntry
	for i := 0; i < len(conf.services); i++ {
		if conf.services[i].sniff == nil {
			nilServices = append(nilServices, conf.services[i])
		} else {
			services = append(services, conf.services[i])
		}
	}
	conf.services = append(services, nilServices...)
	for _, s := range conf.services {
		ln := svr.split(func(rd io.Reader) (n int, ok bool) {
			if s.sniff == nil {
				return 0, true
			}
			return 0, s.sniff(rd)
		})
		go s.serve(newService(m, ra, conf.Auth), ln)
	}
	if conf.InitRunQuit {
		logger.Notice("init run quit")
		os.Exit(0)
	}
}

// Receiver ...
type Receiver interface {
	Recv() (interface{}, time.Duration, error)
}

// SendOptions ...
type SendOptions struct {
	Context        interface{}
	From           interface{}
	AllowOpenReads bool
	DenyOpenReads  bool
}

var defSendOpts = &SendOptions{}

// A Message represents a command and is in a format that is consumed by
// an Observer.
type Message struct {
	// Args are the original command arguments.
	Args []string
	// Resp is the command reponse, if not an error.
	Resp interface{}
	// Err is the command error, if not successful.
	Err error
	// Elapsed is the amount of time that the command took to process.
	Elapsed time.Duration
	// Addr is the remote TCP address of the connection that generated
	// this message.
	Addr string
}

// Service is a client facing service.
type Service interface {
	// Send a command with args from a client
	Send(args []string, opts *SendOptions) Receiver
	// Auth authorizes a client
	Auth(auth string) error
	// Monitor returns a service monitor for observing client commands.
	Monitor() Monitor
	// Opened
	Opened(addr string) (context interface{}, accept bool)
	// Closed
	Closed(context interface{}, addr string)
}

type serviceEntry struct {
	sniff func(rd io.Reader) bool
	serve func(s Service, ln net.Listener)
}

type service struct {
	m    *machine
	ra   *raftWrap
	auth string
	mon  *monitor

	writeMu sync.Mutex
	write   map[interface{}]*writeRequestFuture
}

func newService(m *machine, ra *raftWrap, auth string) *service {
	s := &service{m: m, ra: ra, auth: auth}
	s.write = make(map[interface{}]*writeRequestFuture)
	s.mon = newMonitor(s)
	return s
}

// Monitor allows for observing all incoming service commands from all clients.
// See an example in the examples/kvdb project.
func (s *service) Monitor() Monitor {
	return s.mon
}

func (s *service) Auth(auth string) error {
	if s.auth != auth {
		return ErrUnauthorized
	}
	return nil
}

// The Send function sends command args to the service and return a future
// receiver for getting the response.
// There are three type of commands: write, read, and system.
// - Write commands always go though the raft log one at a time.
// - Read commands do not go though the raft log but do need to be executed
//   on the leader. Many reads from multiple clients can execute at the same
//   time, but each read must wait until the leader has applied at least one
//   new tick (which acts as a barrier) and must wait for any pending writes
//   that the same client has issued to be fully applied before executing the
//   read.
// - System commands run independently from the machine or user data space, and
//   are primarily used for executing lower level system operations such as
//   Raft functions, backups, server stats, etc.
//
// ** Open Reads **
// When the server has been started with the --openreads flag or when
// SendOptions.AllowOpenReads is true, followers can also accept reads.
// Using open reads runs the risk of returning stale data.
func (s *service) Send(args []string, opts *SendOptions) Receiver {
	if len(args) == 0 {
		// Empty command gets an empty response
		return Response(nil, 0, nil)
	}
	cmdName := strings.ToLower(args[0])
	cmd, ok := s.m.commands[cmdName]
	if !ok {
		if s.m.catchall.kind == 0 {
			return Response(nil, 0, ErrUnknownCommand)
		}
		cmd = s.m.catchall
	}
	if cmdName == "tick" {
		// The "tick" command is explicitly denied from being called by a
		// service. It must only be called from the runTicker function.
		// Let's just pretend like it's an unknown command.
		return Response(nil, 0, ErrUnknownCommand)
	}
	if opts == nil {
		// Use the default send options when the sender does not tell us what
		// they want.
		opts = defSendOpts
	}
	switch cmd.kind {
	case 'w': // write
		r := &writeRequestFuture{args: args, s: s, from: opts.From}
		r.wg.Add(1)
		s.m.wrC <- r
		s.addWrite(opts.From, r)
		return r
	case 'r': // read
		s.waitWrite(opts.From)
		start := time.Now()
		resp, err := s.execRead(cmd, args, opts)
		return Response(resp, time.Since(start), errRaftConvert(s.ra, err))
	case 's': // intermediate/system
		s.waitWrite(opts.From)
		start := time.Now()
		pm := intermediateMachine{m: s.m, context: opts.Context}
		resp, err := cmd.fn(pm, s.ra, args)
		return Response(resp, time.Since(start), errRaftConvert(s.ra, err))
	default:
		return Response(nil, 0, errors.New("invalid request"))
	}
}

func (s *service) Opened(addr string) (context interface{}, accept bool) {
	if s.m.connOpened != nil {
		return s.m.connOpened(addr)
	}
	return nil, true
}

func (s *service) Closed(context interface{}, addr string) {
	if s.m.connClosed != nil {
		s.m.connClosed(context, addr)
	}
}

func (s *service) execRead(cmd command, args []string, opts *SendOptions,
) (interface{}, error) {
	openReads := s.m.openReads
	if opts.AllowOpenReads {
		if opts.DenyOpenReads {
			return nil, ErrInvalid
		}
		openReads = true
	} else if opts.DenyOpenReads {
		openReads = false
	}
	var resp interface{}
	var err error
	if openReads {
		resp, err = s.execOpenRead(cmd, args)
	} else {
		resp, err = s.execNonOpenRead(cmd, args)
	}
	return resp, err
}

func (s *service) execOpenRead(cmd command, args []string,
) (interface{}, error) {
	// open reads can be performed on the leaders and followers that have a log
	// which is reasonably loaded.
	if atomic.LoadInt32(&s.m.logLoaded) == 0 {
		return nil, raft.ErrNotLeader
	}
	// Set the machine to read access mode
	s.m.mu.RLock()
	atomic.AddInt32(&s.m.readers, 1)
	defer func() {
		// Return the machine to write access mode
		atomic.AddInt32(&s.m.readers, -1)
		s.m.mu.RUnlock()
	}()
	return cmd.fn(s.m, s.ra, args)
}

func (s *service) execNonOpenRead(cmd command, args []string,
) (interface{}, error) {
	// Non-open reads can only be performed on a leader that has received
	// a tick response. In this case a tick acts as a write barrier ensuring
	// that any read command will always follow the tick.
	s.m.mu.RLock()
	atomic.AddInt32(&s.m.readers, 1)
	defer func() {
		atomic.AddInt32(&s.m.readers, -1)
		s.m.mu.RUnlock()
	}()
	if s.ra.State() != raft.Leader || s.m.tickedIndex == 0 {
		return nil, raft.ErrNotLeader
	}
	// We are the leader and we have received a tick event.
	// Complete the read command.
	return cmd.fn(s.m, s.ra, args)
}

func (s *service) addWrite(from interface{}, r *writeRequestFuture) {
	s.writeMu.Lock()
	s.write[from] = r
	s.writeMu.Unlock()
}

func (s *service) waitWrite(from interface{}) {
	s.writeMu.Lock()
	r := s.write[from]
	s.writeMu.Unlock()
	if r != nil {
		r.Recv()
	}
}

type simpleResponse struct {
	v    interface{}
	elap time.Duration
	err  error
}

func (r *simpleResponse) Recv() (interface{}, time.Duration, error) {
	return r.v, r.elap, r.err
}

// Response ...
func Response(v interface{}, elapsed time.Duration, err error) Receiver {
	return &simpleResponse{v, elapsed, err}
}

// writeRequestFuture is the basic unity of communication from services to the
// raft log. It's a Future type that is sent through a channel, picked up by a
// background routine that then applies the `args` to the raft log. Upon
// successfully being applied, the `resp` is fill with the response, and
// `wg.Done` is called.
type writeRequestFuture struct {
	args []string
	resp interface{}
	err  error
	elap time.Duration
	wg   sync.WaitGroup
	s    *service
	from interface{}
}

// Recv received the response and time elapsed to process the write. Or, it
// returns an error.
func (r *writeRequestFuture) Recv() (interface{}, time.Duration, error) {
	r.wg.Wait()
	r.s.writeMu.Lock()
	if r.s.write[r.from] == r {
		delete(r.s.write, r.from)
	}
	r.s.writeMu.Unlock()
	return r.resp, r.elap, r.err
}
