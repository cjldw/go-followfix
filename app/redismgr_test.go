package app
import (
	"testing"
)

// TEST Redis manager
func TestNewRedisMgr(t *testing.T) {
	redisMgr := NewRedisMgr()
	socialRedis, err := redisMgr.GetRedisByName("social")
	ThrowErr(err)
	t.Log(socialRedis)
}
