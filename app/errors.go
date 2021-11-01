package app

import (
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	"strings"
)

// ErrSyntax is returned where there was a syntax error
var ErrSyntax = errors.New("syntax error")

// ErrNotLeader is returned when the raft leader is unknown
var ErrNotLeader = raft.ErrNotLeader

// ErrWrongNumArgs is returned when the arg count is wrong
var ErrWrongNumArgs = errors.New("wrong number of arguments")

// ErrUnauthorized is returned when a client connection has not been authorized
var ErrUnauthorized = errors.New("unauthorized")

// ErrUnknownCommand is returned when a command is not known
var ErrUnknownCommand = errors.New("unknown command")

// ErrInvalid is returned when an operation has invalid arguments or options
var ErrInvalid = errors.New("invalid")

// ErrCorrupt is returned when a data is invalid or corrupt
var ErrCorrupt = errors.New("corrupt")

var errWrongNumArgsRaft = errors.New("wrong number of arguments, try RAFT HELP")

var errWrongNumArgsCluster = errors.New("wrong number of arguments, " +
	"try CLUSTER HELP")

func errUnknownRaftCommand(args []string) error {
	var cmd string
	for _, arg := range args {
		cmd += arg + " "
	}
	return fmt.Errorf("unknown raft command '%s', try RAFT HELP",
		strings.TrimSpace(cmd))
}

func errUnknownClusterCommand(args []string) error {
	var cmd string
	for _, arg := range args {
		cmd += arg + " "
	}
	return fmt.Errorf("unknown subcommand or wrong number of arguments for "+
		"'%s', try CLUSTER HELP",
		strings.TrimSpace(cmd))
}
