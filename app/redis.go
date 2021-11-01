package app

import (
	"errors"
	"fmt"
	"github.com/moontrade/server/logger"
	"github.com/tidwall/redcon"
	"io"
	"net"
	"os"
	"strings"
)

// redisService provides a service that is compatible with the Redis protocol.
func redisService() (func(io.Reader) bool, func(Service, net.Listener)) {
	return nil, redisServiceHandler
}

type redisClient struct {
	authorized bool
	opts       SendOptions
}

func redisCommandToArgs(cmd redcon.Command) []string {
	args := make([]string, len(cmd.Args))
	args[0] = strings.ToLower(string(cmd.Args[0]))
	for i := 1; i < len(cmd.Args); i++ {
		args[i] = string(cmd.Args[i])
	}
	return args
}

type redisQuitClose struct{}

func redisServiceExecArgs(s Service, client *redisClient, conn redcon.Conn,
	args [][]string,
) {
	recvs := make([]Receiver, len(args))
	var close bool
	for i, args := range args {
		var r Receiver
		switch args[0] {
		case "quit":
			r = Response(redisQuitClose{}, 0, nil)
			close = true
		case "auth":
			if len(args) != 2 {
				r = Response(nil, 0, ErrWrongNumArgs)
			} else if err := s.Auth(args[1]); err != nil {
				client.authorized = false
				r = Response(nil, 0, err)
			} else {
				client.authorized = true
				r = Response(redcon.SimpleString("OK"), 0, nil)
			}
		default:
			if !client.authorized {
				if err := s.Auth(""); err != nil {
					client.authorized = false
					r = Response(nil, 0, err)
				} else {
					client.authorized = true
				}
			}
			if client.authorized {
				switch args[0] {
				case "ping":
					if len(args) == 1 {
						r = Response(redcon.SimpleString("PONG"), 0, nil)
					} else if len(args) == 2 {
						r = Response(args[1], 0, nil)
					} else {
						r = Response(nil, 0, ErrWrongNumArgs)
					}
				case "shutdown":
					logger.Error(errors.New("shutting down"))
					os.Exit(0)
				case "echo":
					if len(args) != 2 {
						r = Response(nil, 0, ErrWrongNumArgs)
					} else {
						r = Response(args[1], 0, nil)
					}
				default:
					r = s.Send(args, &client.opts)
				}
			}
		}
		recvs[i] = r
		if close {
			break
		}
	}
	// receive responses
	var filteredArgs [][]string
	for i, r := range recvs {
		resp, elapsed, err := r.Recv()
		if err != nil {
			if err == ErrUnknownCommand {
				err = fmt.Errorf("%s '%s'", err, args[i][0])
			}
			conn.WriteAny(err)
		} else {
			switch v := resp.(type) {
			case FilterArgs:
				filteredArgs = append(filteredArgs, v)
			case Hijack:
				conn := newRedisHijackedConn(conn.Detach())
				go v(s, conn)
			case redisQuitClose:
				conn.WriteString("OK")
				conn.Close()
			default:
				conn.WriteAny(v)
			}
		}
		// broadcast the request and response to all observers
		s.Monitor().Send(Message{
			Addr:    conn.RemoteAddr(),
			Args:    args[i],
			Resp:    resp,
			Err:     err,
			Elapsed: elapsed,
		})
	}
	if len(filteredArgs) > 0 {
		redisServiceExecArgs(s, client, conn, filteredArgs)
	}
}

func redisServiceHandler(s Service, ln net.Listener) {
	logger.Fatal(redcon.Serve(ln,
		// handle commands
		func(conn redcon.Conn, cmd redcon.Command) {
			client := conn.Context().(*redisClient)
			var args [][]string
			args = append(args, redisCommandToArgs(cmd))
			for _, cmd := range conn.ReadPipeline() {
				args = append(args, redisCommandToArgs(cmd))
			}
			redisServiceExecArgs(s, client, conn, args)
		},
		// handle opened connection
		func(conn redcon.Conn) bool {
			context, accept := s.Opened(conn.RemoteAddr())
			if !accept {
				return false
			}
			client := new(redisClient)
			client.opts.From = client
			client.opts.Context = context
			conn.SetContext(client)
			return true
		},
		// handle closed connection
		func(conn redcon.Conn, err error) {
			if conn.Context() == nil {
				return
			}
			client := conn.Context().(*redisClient)
			s.Closed(client.opts.Context, conn.RemoteAddr())
		}),
	)
}

// FilterArgs ...
type FilterArgs []string

// Hijack is a function type that can be used to "hijack" a service client
// connection and allowing to perform I/O operations outside the standard
// network loop. An example of it's usage can be found in the examples/kvdb
// project.
type Hijack func(s Service, conn HijackedConn)

// HijackedConn is a connection that has been detached from the main service
// network loop. It's entirely up to the hijacker to performs all I/O
// operations. The Write* functions buffer write data and the Flush must be
// called to do the actual sending of the data to the connection.
// Close the connection to when done.
type HijackedConn interface {
	// RemoteAddr is the connection remote tcp address.
	RemoteAddr() string
	// ReadCommands is an iterator function that reads pipelined commands.
	// Returns a error when the connection encountared and error.
	ReadCommands(func(args []string) bool) error
	// ReadCommand reads one command at a time.
	ReadCommand() (args []string, err error)
	// WriteAny writes any type to the write buffer using the format rules that
	// are defined by the original Service.
	WriteAny(v interface{})
	// WriteRaw writes raw data to the write buffer.
	WriteRaw(data []byte)
	// Flush the write write buffer and send data to the connection.
	Flush() error
	// Close the connection
	Close() error
}

type redisHijackConn struct {
	dconn redcon.DetachedConn
	cmds  []redcon.Command
}

func newRedisHijackedConn(dconn redcon.DetachedConn) *redisHijackConn {
	return &redisHijackConn{dconn: dconn}
}

func (conn *redisHijackConn) ReadCommands(iter func(args []string) bool) error {
	if len(conn.cmds) == 0 {
		cmd, err := conn.dconn.ReadCommand()
		if err != nil {
			return err
		}
		if !iter(redisCommandToArgs(cmd)) {
			return nil
		}
		conn.cmds = conn.dconn.ReadPipeline()
	}
	for len(conn.cmds) > 0 {
		cmd := conn.cmds[0]
		conn.cmds = conn.cmds[1:]
		if !iter(redisCommandToArgs(cmd)) {
			return nil
		}
	}
	return nil
}

func (conn *redisHijackConn) WriteAny(v interface{}) {
	conn.dconn.WriteAny(v)
}

func (conn *redisHijackConn) WriteRaw(data []byte) {
	conn.dconn.WriteRaw(data)
}

func (conn *redisHijackConn) Flush() error {
	return conn.dconn.Flush()
}

func (conn *redisHijackConn) Close() error {
	return conn.dconn.Close()
}

func (conn *redisHijackConn) RemoteAddr() string {
	return conn.dconn.RemoteAddr()
}

func (conn *redisHijackConn) ReadCommand() (args []string, err error) {
	err = conn.ReadCommands(func(iargs []string) bool {
		args = iargs
		return false
	})
	return args, err
}
