package app

import (
	"testing"
	log "github.com/cihub/seelog"
)

func TestInitLog(t *testing.T)  {
	InitLog("../conf/log4go.xml")
	log.Error("ooooooooooooooo")
	log.Info("hello world")
}
