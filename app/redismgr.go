package app
import (
	"github.com/go-redis/redis"
	"strconv"
	"errors"
	"fmt"
	//"log"
	"sync"
)

type RedisMgr struct {
	redisList map[string]*redis.Client
	Lock *sync.RWMutex
}

// NewRedisMgr create redisMgr object
func NewRedisMgr() *RedisMgr {
	return &RedisMgr{redisList:make(map[string]*redis.Client), Lock:&sync.RWMutex{}}
}

// InitializeRedisList
func (redisMgr *RedisMgr) InitializeRedisList (redisConfMap map[string]RedisInst) {
	for redisKey, redisConf := range redisConfMap {
		if _, ok := redisMgr.redisList[redisKey]; ok {
			continue
		}
		var redisOpt redis.Options
		if redisConf.Socket != "" {
			redisOpt = redis.Options{
				Network:"unix",
				Addr: redisConf.Socket,
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
		redisMgr.redisList[redisKey] = redisClient
	}
}

// GetRedisInstByName
// get redis instance by redis configure
func (redisMgr *RedisMgr) GetRedisByName(redisKey string) (*redis.Client , error) {
	redisClient, ok := redisMgr.redisList[redisKey]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Redis Key 【%s】not exists！", redisKey))
	}
	/*
	err := redisClient.Ping().Err()
	if err != nil { // reconnect redis
		redisClient.Close()
		log.Println("Reconnect Redis")
		redisConfMap := GetApp().confmgr.RedisConf
		redisConf, ok := redisConfMap[redisKey]
		if !ok {
			return nil, errors.New(fmt.Sprintf("Redis Key 【%s】not exists！", redisKey))
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
		redisClient = redis.NewClient(&redisOpt)
		redisList[redisKey] = redisClient
	}
	*/
	return redisClient, nil
}