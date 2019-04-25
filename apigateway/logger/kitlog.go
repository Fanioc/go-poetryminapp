package logger

import (
	"github.com/go-kit/kit/log"
	"os"
)

func CreateKitLog() log.Logger {
	logger := log.NewLogfmtLogger(os.Stderr) // try to openfile
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
	return logger
}
