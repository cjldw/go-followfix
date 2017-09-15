package go_followfix

import (
	"testing"
)

var appConf *AppConf
func initializeConf()  {
	var err error
	appConf, err = NewAppConf().Load("./conf/app.toml", false)
	ThrowErr(err)
}

// TEST Redis manager
func TestNewRedisMgr(t *testing.T) {
	initializeConf()
	redisMgr := NewRedisMgr()
	t.Log(appConf.RedisConf)
	redisMgr.InitializeRedisList(appConf.RedisConf)
	socialRedis, err := redisMgr.GetRedisInstByName("social")
	ThrowErr(err)
	t.Log(socialRedis)
}
