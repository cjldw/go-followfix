package app
import (
	"testing"
	"github.com/go-redis/redis"
)

// TEST Redis manager
func TestNewRedisMgr(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr: "127.0.0.1:6379",
		DB: 9,
	})

	//redisClient.HSet("ol", "12", "luowen")

	//err := redisClient.HExists("ol", "12")
	v, e := redisClient.HGet("ol", "12").Result()
	t.Log(v=="")
	t.Log("----------")
	t.Log(e == nil)
}
