package log

import (
	"os"

	"github.com/op/go-logging"
)

var (
	log    = logging.MustGetLogger("gitlab-cli")
	format = logging.MustStringFormatter(`%{color}%{level:.4s} ▶%{color:reset} %{message}`)
)

func Setup(level string) {
	b := logging.NewLogBackend(os.Stdout, "", 0)
	bformatter := logging.NewBackendFormatter(b, format)
	logging.SetBackend(bformatter)

	logging.AddModuleLevel(b)

	loglevel, err := logging.LogLevel(level)
	if err != nil {
		log.Error("unrecognised log level, using info")
		loglevel = logging.INFO
	}
	logging.SetLevel(loglevel, "")
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}
