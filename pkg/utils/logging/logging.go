package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.Out = os.Stdout
	// hardcoding log level to debug for now
	log.SetLevel(logrus.DebugLevel)
}

func GetLogger() *logrus.Logger {
	return log
}
