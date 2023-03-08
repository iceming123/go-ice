package testlog

import "github.com/iceming123/go-ice/log"

var msg string = "P2P"

func AddLog(ctx ...interface{}) {
	log.Info(msg, ctx...)
}
