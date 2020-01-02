package blog

import (
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.SetLevel(logrus.DebugLevel)
}
