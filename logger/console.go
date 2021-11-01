package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"strings"
)

const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite

	colorBold     = 1
	colorDarkGray = 90
)

func SetConsoleWriter() {
	log = zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = os.Stderr
		w.FormatLevel = consoleDefaultFormatLevel(false)
		w.TimeFormat = "15:04:05.000"
	}))
}

func SetJsonWriter() {
	log = zerolog.New(os.Stderr)
}

// colorize returns the string s wrapped in ANSI code c, unless disabled is true.
func colorize(s interface{}, c int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}

func consoleDefaultFormatLevel(noColor bool) zerolog.Formatter {
	return func(i interface{}) string {
		var l string
		if ll, ok := i.(string); ok {
			switch strings.ToLower(ll) {
			case "trace":
				l = colorize("TRC", colorMagenta, noColor)
			case "debug":
				l = colorize("DBG", colorYellow, noColor)
			case "notice":
				l = colorize("NOT", colorYellow, noColor)
			case "info":
				l = colorize("INF", colorGreen, noColor)
			case "warn":
				l = colorize("WRN", colorRed, noColor)
			case "alert":
				l = colorize(colorize("ALT", colorRed, noColor), colorBold, noColor)
			case "error":
				l = colorize(colorize("ERR", colorRed, noColor), colorBold, noColor)
			case "fatal":
				l = colorize(colorize("FTL", colorRed, noColor), colorBold, noColor)
			case "panic":
				l = colorize(colorize("PNC", colorRed, noColor), colorBold, noColor)
			case "emergency":
				l = colorize(colorize("FTL", colorRed, noColor), colorBold, noColor)
			default:
				l = colorize("???", colorBold, noColor)
			}
		} else {
			if i == nil {
				l = colorize("???", colorBold, noColor)
			} else {
				l = strings.ToUpper(fmt.Sprintf("%s", i))[0:3]
			}
		}
		return l
	}
}
