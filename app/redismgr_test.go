package app
import (
	"testing"
)

// TEST Redis manager
func TestNewRedisMgr(t *testing.T) {
	InitializeConf()
	redisMgr := NewRedisMgr()
	t.Log(appConf.RedisConf)
	redisMgr.InitializeRedisList(appConf.RedisConf)
	socialRedis, err := redisMgr.GetRedisByName("social")
	ThrowErr(err)
	t.Log(socialRedis)
}
