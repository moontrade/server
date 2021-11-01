package app

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/moontrade/server/logger"
	"github.com/tidwall/redlog/v2"
	"os"
	"strings"
)

func logInit(conf Config) hclog.Logger {
	level := hclog.Error
	switch conf.LogLevel {
	case "debug":
		level = hclog.Debug
	case "verbose", "verb":
		level = hclog.Trace
	case "notice", "info":
		level = hclog.Info
	case "warning", "warn":
		level = hclog.Warn
	case "quiet", "silent":
		level = hclog.NoLevel
		//wr = ioutil.Discard
	default:
		fmt.Fprintf(os.Stderr, "invalid -loglevel: %s\n", conf.LogLevel)
		os.Exit(1)
	}
	//logger.SetWriter(wr)
	logger.SetConsoleWriter()
	hclopts := *hclog.DefaultOptions
	hclopts.Level = level
	//hclopts.Color = hclog.ColorOff
	//hclopts.Output = logger
	hclopts.Output = logger.RaftWriter
	logger.Warn("starting %s", versline(conf))
	return hclog.New(&hclopts)
}

func stateChangeFilter(line string, log *redlog.Logger) string {
	if strings.Contains(line, "entering ") {
		app := log.App()
		if strings.Contains(line, "entering candidate state") {
			app = 'C'
		} else if strings.Contains(line, "entering follower state") {
			app = 'F'
		} else if strings.Contains(line, "entering leader state") {
			app = 'L'
		} else {
			return line
		}
		log.SetApp(app)
	}
	return line
}
