package logger

import (
	"github.com/rs/zerolog"
	"strings"
	"unsafe"
)

type raftWriter struct {
	level zerolog.Level
	l     zerolog.Logger
}

var (
	RaftWriter = &raftWriter{}
)

// Write writes to the log
func (l *raftWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(*(*string)(unsafe.Pointer(&p)))
	idx := strings.IndexByte(msg, ' ')
	if idx != -1 {
		msg = msg[idx+1:]
	}
	idx = strings.IndexByte(msg, ']')
	var level zerolog.Level
	if idx != -1 && msg[0] == '[' {
		switch msg[1] {
		default: // -> verbose
			level = zerolog.DebugLevel
		case 'W': // warning -> warning
			level = zerolog.WarnLevel
		case 'E': // error -> warning
			level = zerolog.ErrorLevel
		case 'D': // debug -> debug
			level = zerolog.DebugLevel
		case 'V': // verbose -> verbose
			level = zerolog.TraceLevel
		case 'I': // info -> notice
			level = zerolog.InfoLevel
		}
		msg = msg[idx+1:]
		for len(msg) > 0 && msg[0] == ' ' {
			msg = msg[1:]
		}
	}

	//if tty {
	//	msg = strings.Replace(msg, "[Leader]",
	//		"\x1b[32m[Leader]\x1b[0m", 1)
	//	msg = strings.Replace(msg, "[Follower]",
	//		"\x1b[33m[Follower]\x1b[0m", 1)
	//	msg = strings.Replace(msg, "[Candidate]",
	//		"\x1b[36m[Candidate]\x1b[0m", 1)
	//}
	idx = strings.IndexByte(msg, ':')
	var _args [16]interface{}
	args := _args[:0]

	if idx > -1 {
		fields := strings.TrimSpace(msg[idx+1:])
		msg = strings.TrimSpace(msg[0:idx])

		for len(fields) > 0 {
			idx = strings.IndexByte(fields, '=')
			if idx == -1 {
				args = append(args, fields)
				args = append(args, "")
				break
			}
			name := strings.TrimSpace(fields[0:idx])
			fields = strings.TrimSpace(fields[idx+1:])
			if len(fields) == 0 {
				args = append(args, name)
				args = append(args, "")
				break
			}

			if fields[0] == '"' {
				fields = fields[1:]
				idx = strings.IndexByte(fields, '"')
				if idx == -1 {
					args = append(args, name)
					args = append(args, fields)
					break
				}
				args = append(args, name)
				args = append(args, fields[0:idx])
				fields = strings.TrimSpace(fields[idx+1:])
			} else {
				idx = strings.IndexByte(fields, ' ')
				if idx == -1 {
					args = append(args, name)
					args = append(args, fields)
					break
				}
				args = append(args, name)
				args = append(args, fields[0:idx])
				fields = strings.TrimSpace(fields[idx+1:])
			}

			if len(args) > 13 {
				break
			}
		}

		args = append(args, msg)
	}

	if len(args) > 0 {
		DoCaller(level, 6, args...)
	} else {
		DoCaller(level, 6, msg)
	}
	return len(p), nil
}
