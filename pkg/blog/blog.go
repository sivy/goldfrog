package blog

import (
	"time"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func init() {
	logger.SetLevel(logrus.DebugLevel)
}

const (
	POSTDATEFMT      string = "2006-01-02"
	POSTTIMESTAMPFMT string = time.RFC3339
)
