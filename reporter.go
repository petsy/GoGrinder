package gogrinder

import (
	log "github.com/Sirupsen/logrus"
	time "github.com/finklabs/ttime"
)

// event logger instance
var elog *log.Logger

func eventLogger(meta Meta) {
	elog.WithFields(log.Fields{
		// TODO: timestamp needs proper formatting
		"timestamp": meta["timestamp"].(time.Time),
		"user":      meta["user"].(int),
		"iteration": meta["iteration"].(int),
		"testcase":  meta["testcase"].(string),
		"elapsed":   meta["elapsed"].(time.Duration),
	}).Info()
}
