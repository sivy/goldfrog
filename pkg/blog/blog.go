package blog

import (
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func init() {
	logger.SetLevel(logrus.DebugLevel)
}

const (
	POSTDATEFMT      string = "2006-01-02"
	POSTTIMESTAMPFMT string = "2006-01-02 15:04"
)
