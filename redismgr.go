package go_followfix

import (
	"github.com/go-redis/redis"
	"strconv"
	"errors"
	"fmt"
)

type RedisMgr struct {
	RedisClient *redis.Client
}

var redisList map[string]*redis.Client = make(map[string]*redis.Client)

func NewRedisMgr() *RedisMgr {
	return &RedisMgr{}
}

// InitializeRedisList
func (RedisMgr *RedisMgr) InitializeRedisList (redisConfMap map[string]RedisInst) {
	fmt.Println("-----------------------")
	for redisKey, redisConf := range redisConfMap {
		if _, ok := redisList[redisKey]; ok {
			continue
		}
		var redisOpt redis.Options
		if redisConf.Socket != "" {
			redisOpt = redis.Options{
				Network: "unix://"+redisConf.Socket,
				DB: redisConf.Db,
			}
		} else {
			redisOpt = redis.Options{
				Addr: redisConf.Host + ":" + strconv.Itoa(redisConf.Port),
				Password: redisConf.Password,
				DB: redisConf.Db,
			}
		}
		redisClient := redis.NewClient(&redisOpt)
		redisList[redisKey] = redisClient
	}
}

// GetRedisInstByName
// get redis instance by redis configure
func (redisMgr *RedisMgr)GetRedisInstByName(redisKey string) (*RedisMgr , error) {
	redisClient, ok := redisList[redisKey]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Redis Key 【%s】not exists！", redisKey))
	}
	 err := redisClient.Ping().Err()
	ThrowErr(err)
	redisMgr.RedisClient = redisClient
	return redisMgr, nil
}