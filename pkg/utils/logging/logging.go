package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.Out = os.Stdout
}

func GetLogger() *logrus.Logger {
	return log
}
