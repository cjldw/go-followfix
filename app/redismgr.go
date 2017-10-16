package app
import (
	"github.com/go-redis/redis"
	"strconv"
	"errors"
	"fmt"
	//"log"
)

type RedisMgr struct {
	RedisClient *redis.Client
}

// redis list container
var redisList map[string]*redis.Client = make(map[string]*redis.Client)

// NewRedisMgr create redisMgr object
func NewRedisMgr() *RedisMgr {
	return &RedisMgr{}
}

// InitializeRedisList
func (RedisMgr *RedisMgr) InitializeRedisList (redisConfMap map[string]RedisInst) {
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
func (redisMgr *RedisMgr) GetRedisByName(redisKey string) (*RedisMgr , error) {
	redisClient, ok := redisList[redisKey]
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
	redisMgr.RedisClient = redisClient
	return redisMgr, nil
}