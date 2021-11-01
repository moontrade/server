package logger

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

var (
	log zerolog.Logger

	DurationAsString   = true
	RawFieldName       = "raw"
	DataFieldName      = "data"
	DurationFieldName  = "dur"
	DurationsFieldName = "durs"
	ErrorsFieldName    = "errors"

	EmptyMessage = ""
)

const (
	RequestName  = "request"
	ResponseName = "response"
)

func Log() *zerolog.Logger {
	return &log
}

//type Logger interface {
//	Debugf(format string, args ...interface{})
//	Debug(args ...interface{})
//	Debugln(args ...interface{})
//	Verbf(format string, args ...interface{})
//	Verb(args ...interface{})
//	Verbln(args ...interface{})
//	Noticef(format string, args ...interface{})
//	Notice(args ...interface{})
//	Noticeln(args ...interface{})
//	Printf(format string, args ...interface{})
//	Print(args ...interface{})
//	Println(args ...interface{})
//	Warningf(format string, args ...interface{})
//	Warning(args ...interface{})
//	Warningln(args ...interface{})
//	Fatalf(format string, args ...interface{})
//	Fatal(args ...interface{})
//	Fatalln(args ...interface{})
//	Panicf(format string, args ...interface{})
//	Panic(args ...interface{})
//	Panicln(args ...interface{})
//	Errorf(format string, args ...interface{})
//	Error(args ...interface{})
//	Errorln(args ...interface{})
//}

// JSON Tag a byte slice as being a JSON document
type JSON []byte

type Request interface{}
type Response interface{}

type Builder func(event *zerolog.Event)

func init() {
	setCallerFormatter()

	// Use GCP cloud logging naming
	zerolog.LevelFieldName = "severity"
	zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
		switch l {
		case zerolog.TraceLevel:
			return "DEFAULT"
		case zerolog.DebugLevel:
			return "DEBUG"
		case zerolog.InfoLevel:
			return "INFO"
		case zerolog.NoLevel:
			return "NOTICE"
		case zerolog.WarnLevel:
			return "WARN"
		case zerolog.ErrorLevel:
			return "ERROR"
		case zerolog.PanicLevel:
			return "CRITICAL"
		case zerolog.FatalLevel:
			return "EMERGENCY"
		default:
			return "DEFAULT"
		}
	}

	// Default to Trace level
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	// Default output to console
	SetConsoleWriter()
}

func setCallerFormatter() {
	_, file, _, _ := runtime.Caller(0)
	prefix := path.Dir(path.Dir(file))
	if len(prefix) > 0 && prefix[len(prefix)-1] != os.PathSeparator {
		prefix += "/"
	}

	zerolog.CallerMarshalFunc = func(file string, line int) string {
		if prefix == "" {
			return fmt.Sprintf("%s:%d", file, line)
		}
		index := strings.Index(file, prefix)
		if index > -1 {
			file = file[index+len(prefix):]
		}
		return fmt.Sprintf("%s:%d", file, line)
	}
}

func SetWriter(w io.Writer) {
	log = zerolog.New(w)
}

func SetLogger(logger zerolog.Logger) {
	log = logger
}

func appendInterface(event *zerolog.Event, name string, value interface{}) {
	switch v := value.(type) {
	case time.Duration:
		if DurationAsString {
			event.Str(name, v.String())
		} else {
			event.Dur(name, v)
		}
	case JSON:
		event.RawJSON(name, v)
	case json.Marshaler:
		bytes, _ := v.MarshalJSON()
		event.RawJSON(name, bytes)
	default:
		event.Interface(name, value)
	}
}

func doLog(skip int, event *zerolog.Event, args []interface{}) {
	// Timestamp
	event.Timestamp()
	// Caller
	event.Caller(skip)

	// Builder?
	if len(args) == 0 {
		event.Msg(EmptyMessage)
		return
	}

	//// Treat as simple message?
	//if len(args) == 1 {
	//	arg0 := args[0]
	//	msg, ok := arg0.(string)
	//	if !ok {
	//		msg = fmt.Sprintf("%s", arg0)
	//	}
	//	event.Msg(msg)
	//	return
	//}

	switch t := args[0].(type) {
	// Handle error
	case error:
		event.Err(t)
		args = args[1:]
	}

	for i := 0; i < len(args); i += 2 {
		key := args[i]
		if key == nil {
			// Shift by one
			i -= 1
			continue
		}

		switch k := key.(type) {
		case string:
			// Treat it like a format template?
			if strings.Contains(k, "%") {
				// The remaining args will be the format values
				event.Msgf(k, args[i+1:]...)
				return
			}

			valueIndex := i + 1
			// Treat key as message?
			if valueIndex == len(args) {
				event.Msg(k)
				return
			}

			// Add to field map
			value := args[valueIndex]
			switch v := value.(type) {
			case string:
				event.Str(k, v)
			case time.Time:
				event.Time(k, v)
			case *time.Time:
				event.Time(k, *v)
			case int:
				event.Int(k, v)
			case int8:
				event.Int8(k, v)
			case int16:
				event.Int16(k, v)
			case int32:
				event.Int32(k, v)
			case int64:
				event.Int64(k, v)
			case uint:
				event.Uint(k, v)
			case uint8:
				event.Uint8(k, v)
			case uint16:
				event.Uint16(k, v)
			case uint32:
				event.Uint32(k, v)
			case uint64:
				event.Uint64(k, v)
			case float32:
				event.Float32(k, v)
			case float64:
				event.Float64(k, v)
			case bool:
				event.Bool(k, v)
			case error:
				event.AnErr(k, v)
			case time.Duration:
				if DurationAsString {
					event.Str(k, v.String())
				} else {
					event.Dur(k, v)
				}
			case json.Marshaler:
				bytes, err := v.MarshalJSON()
				if err != nil {
					event.AnErr(k, err)
				} else {
					event.RawJSON(k, bytes)
				}
			case JSON:
				event.RawJSON(k, v)
			case Builder:
				v(event)
			default:
				appendInterface(event, k, v)
			}
			continue

		case error:
			event.Err(k)
		case []error:
			event.Errs(ErrorsFieldName, k)
		case time.Duration:
			if DurationAsString {
				event.Str(DurationFieldName, k.String())
			} else {
				event.Dur(DurationFieldName, k)
			}
		case []time.Duration:
			event.Durs(DurationsFieldName, k)
		case json.Marshaler:
			bytes, err := k.MarshalJSON()
			if err != nil {
				event.AnErr(DataFieldName, err)
			} else {
				event.RawJSON(DataFieldName, bytes)
			}
		case []byte:
			event.Bytes(RawFieldName, k)
		case JSON:
			event.RawJSON(DataFieldName, k)
		case Builder:
			k(event)
		case Request:
			appendInterface(event, RequestName, k)
		case Response:
			appendInterface(event, ResponseName, k)
		}
		i -= 1
	}

	event.Msg(EmptyMessage)
}

func withErr(event *zerolog.Event, err error) *zerolog.Event {
	if err != nil {
		event.Err(err)
	}
	return event
}

func CustomLevel(level string) *zerolog.Event {
	l := log.Level(zerolog.NoLevel)
	return l.Log().Str(zerolog.LevelFieldName, level)
}

func Do(level zerolog.Level, args ...interface{}) {
	doLog(2, log.WithLevel(level), args)
}

func DoCaller(level zerolog.Level, skip int, args ...interface{}) {
	doLog(skip, log.WithLevel(level), args)
}

// Trace logs a message at level Trace on the standard logger.
func Trace(args ...interface{}) {
	doLog(2, log.Trace(), args)
}

func TraceCaller(skip int, args ...interface{}) {
	doLog(skip, log.Trace(), args)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	doLog(2, log.Debug(), args)
}

func DebugEvent() *zerolog.Event {
	return log.Debug()
}

func DebugCaller(skip int, args ...interface{}) {
	doLog(skip, log.Debug(), args)
}

// Print logs a message at level Info on the standard logger.
func Print(args ...interface{}) {
	doLog(2, log.Info(), args)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	doLog(2, log.Info(), args)
}

func InfoCaller(skip int, args ...interface{}) {
	doLog(skip, log.Info(), args)
}

// Notice logs a message at level Notice on the standard logger.
func Notice(args ...interface{}) {
	doLog(2, CustomLevel("NOTICE"), args)
}

// NoticeCaller logs a message at level Notice on the standard logger.
func NoticeCaller(skip int, args ...interface{}) {
	doLog(skip, CustomLevel("NOTICE"), args)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	doLog(2, log.Warn(), args)
}

func WarnCaller(skip int, args ...interface{}) {
	doLog(skip, log.Warn(), args)
}

// WarnErr logs a message at level Error on the standard logger.
func WarnErr(err error, args ...interface{}) {
	doLog(2, log.Warn().Err(err), args)
}

// WarnErrCaller logs a message at level Error on the standard logger.
func WarnErrCaller(err error, skip int, args ...interface{}) {
	doLog(skip, log.Warn().Err(err), args)
}

// Error logs a message at level Error on the standard logger.
func Error(err error, args ...interface{}) {
	doLog(2, log.Error().Err(err), args)
}

func ErrorCaller(skip int, err error, args ...interface{}) {
	doLog(skip, log.Error().Err(err), args)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(err error, args ...interface{}) {
	doLog(2, log.Panic().Err(err), args)
}

func PanicCaller(skip int, err error, args ...interface{}) {
	doLog(skip, log.Panic().Err(err), args)
}

// Alert logs a message at level Panic on the standard logger.
func Alert(args ...interface{}) {
	doLog(2, CustomLevel("ALERT"), args)
}

func AlertCaller(skip int, args ...interface{}) {
	doLog(skip, CustomLevel("ALERT"), args)
}

// Fatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatal(err error, args ...interface{}) {
	doLog(2, log.Fatal().Err(err), args)
}

func FatalCaller(skip int, err error, args ...interface{}) {
	doLog(skip, log.Fatal().Err(err), args)
}
