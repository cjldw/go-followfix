package app

import (
	"sync"
	set "gopkg.in/fatih/set.v0"
	"fmt"
	"strconv"
	"log"
	"github.com/go-redis/redis"
)

var followService *FollowService

type FollowService struct {
	excludeUIDSet *set.Set
	excludeAnchorIdSet *set.Set
	Lock *sync.RWMutex
	Traffic chan PeerUID
	ProduceEnd chan bool
}

type PeerUID struct {
	UID int
	FollowCnt *set.Set
	FansCnt *set.Set
	FriendsCnt *set.Set
}

var followOnce *sync.Once = &sync.Once{}

func NewFollowService() *FollowService {
	if followService == nil {
		followOnce.Do(func() {
			followService = &FollowService{set.New(), set.New(), &sync.RWMutex{}, make(chan PeerUID, 20), make(chan bool)}
		})
	}
	return followService
}

func (followService *FollowService) Produce()  {
	wg := &sync.WaitGroup{}
	for table := 0; table < USER_FOLLOW_SPLIT_TABLE_NUM ; table++  {
		tablename := USER_FOLLOW_TABLE_PREFIX + strconv.Itoa(table)
		wg.Add(1)
		go (func() {
			followService.processSplitTable(tablename)
			wg.Done()
		})()
	}
	wg.Wait()
	followService.ProduceEnd <- true // mark as produce end
}

func (followService *FollowService) Consumer()  {
	wg := &sync.WaitGroup{}
	lock := &sync.RWMutex{}
	produceEnd := false
	for  {
		select {
		case flag := <- followService.ProduceEnd:
			produceEnd = flag
		case peerUID := <- followService.Traffic:
			wg.Add(1)
			go (func() {
				followService.WriteDbRedis(peerUID, lock)
				wg.Done()
			})()
		}
		if produceEnd {
			break
		}
	}
	wg.Wait()
	log.Println("所有用户数据处理完毕")
}

// WriteDbRedis 将单个UID用户写入到Reids中, 更新数据库
func (followService *FollowService) WriteDbRedis( peerUID PeerUID, lock *sync.RWMutex)  {
	uid := peerUID.UID
	// fansCnt := peerUID.FansCnt.Size()

	// lock.Lock()
	// defer lock.Unlock()
	log.Printf("Write to [%d] to redis", uid)

	/*
	dbUsersData, err := GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
	CheckErr(err)
	tableName := USER_FOLLOW_TABLE_PREFIX + strconv.Itoa(uid % USER_FOLLOW_SPLIT_TABLE_NUM)
	fansCntSql := fmt.Sprintf("update %s set fansCnt = %d where uid = %d", tableName, fansCnt, uid)
	log.Println(fansCntSql)
	affected, err := dbUsersData.Db.Exec(fansCntSql)
	CheckErr(err)
	affectedRows, err := affected.RowsAffected()
	CheckErr(err)
	log.Println(fmt.Sprintf("UPDATE FANS OK  ROW 【%d】", affectedRows))

	for index := 0; index < USER_FOLLOW_SPLIT_TABLE_NUM ; index++  { // can not execute multiple statement
		anchorCntSql := fmt.Sprintf("update user_follow_%d set anchorFansCnt = %d where anchor = %d;", index, fansCnt, uid)
		affected, err = dbUsersData.Db.Exec(anchorCntSql)
		log.Println(anchorCntSql)
		CheckErr(err)
		affectedRows, err = affected.RowsAffected()
		CheckErr(err)
		log.Println(fmt.Sprintf("UPDATE ANCHOR OK ROW 【%d】", affectedRows))
	}
	*/

	redisSocial, err := GetApp().redismgr.GetRedisByName(REDIS_SOCIAL)
	CheckErr(err)
	var followListKey string = fmt.Sprintf("%s%d", FRIEND_SYSTEM_USER_FOLLOW, uid)
	var fansListKey string = fmt.Sprintf("%s%d", FRIEND_SYSTEM_USER_FANS, uid)
	var friendsListKey string = fmt.Sprintf("%s%d", FRIEND_SYSTEM_USER_FRIENDS, uid)

	// Fetch UID's Follow List And Storage To Social Redis
	for  { // 处理关注
		followUID := peerUID.FollowCnt.Pop()
		log.Printf("用户[%d]关注 ->［%v］", uid, followUID)
		if followUID == nil {
			break
		}
		iFollowUID, ok := followUID.(int)
		if !ok {
			continue
		}
		followUIDFansCnt := getUIDFansCnt(iFollowUID)
		log.Printf("用户[%d]粉丝 ->［%v］", uid, followUID)
		item := redis.Z{
			Score: float64(followUIDFansCnt),
			Member: followUID,
		}
		redisSocial.RedisClient.ZAdd(followListKey, item)
	}

	for { // 处理粉丝
		fansUID := peerUID.FansCnt.Pop()
		if fansUID == nil {
			break
		}
		iFansUID, ok := fansUID.(int)
		if !ok {
			continue
		}
		fansUIDFansCnt := getUIDFansCnt(iFansUID)
		item := redis.Z{
			Score: float64(fansUIDFansCnt),
			Member: fansUID,
		}
		redisSocial.RedisClient.ZAdd(fansListKey, item)
	}

	for { // 处理好友
		friendsUID := peerUID.FriendsCnt.Pop()
		if friendsUID == nil {
			break
		}
		iFriendUID, ok := friendsUID.(int)
		if !ok {
			continue
		}
		friendsUIDFansCnt := getUIDFansCnt(iFriendUID)
		item := redis.Z{
			Score: float64(friendsUIDFansCnt),
			Member: friendsUID,
		}
		redisSocial.RedisClient.ZAdd(friendsListKey, item)
	}

}

// getUIDFansCnt get user's fans number
// if redis cache return direct
// or select from MySQL then save and return
func getUIDFansCnt(uid int) int {
	redisSocial, err := GetApp().redismgr.GetRedisByName(REDIS_SOCIAL)
	CheckErr(err)
	fansCntKey := fmt.Sprintf("id_%d_fanscnt", uid)
	log.Println(fansCntKey)
	sFansCnt, err := redisSocial.RedisClient.HGet(TMP_UID_FANS_NUM, fansCntKey).Result()
	log.Println(sFansCnt)
	if err == nil {
		iFansCnt, err := strconv.Atoi(sFansCnt)
		if err != nil {
			log.Printf("Convert fans count [%s] ERROR!\n", sFansCnt)
			return 0
		}
		return iFansCnt
	}
	var sql string
	dbUserData, err := GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
	CheckErr(err)
	var fansCnt, fragmentCnt int
	for index := 0; index < USER_FOLLOW_SPLIT_TABLE_NUM ; index++ {
		sql = fmt.Sprintf("select count(*) as fansCnt from user_follow_%d where anchor = %d and status = 1 " +
			"and isFriends = 0", index, uid)
		log.Println("查询粉丝数量:" + sql)
		dbUserData.Db.QueryRow(sql).Scan(&fragmentCnt)
		fansCnt += fragmentCnt
	}
	redisSocial.RedisClient.HSet(TMP_UID_FANS_NUM, fansCntKey, strconv.Itoa(fansCnt))
	return fansCnt
}

// processSplitTable 处理分表数据
func (followService *FollowService) processSplitTable(tablename string)  {
	dbUsersData, err := GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
	CheckErr(err)
	sql := fmt.Sprintf("select uid, anchor from %s where isFriends = 0", tablename)
	/*
	if !followService.excludeUIDSet.IsEmpty() { // exclude has process uid
		uidList1 := followService.excludeUIDSet.List()
		uidList2 := make([]string, len(uidList1))
		for index := range uidList1 {
			uidList2[index] = strconv.Itoa(uidList1[index].(int))
		}
		//sql = fmt.Sprintf("%s and uid not in (%s)", sql, strings.Join(uidList2, ","))
		sql += fmt.Sprintf(" and uid not in (%s)", strings.Join(uidList2, ","))
	}
	if !followService.excludeAnchorIdSet.IsEmpty() { // exclude has process anchor

		anchorIdList := followService.excludeAnchorIdSet.List()
		anchorIdList2 := make([]string, len(anchorIdList))
		for index := range anchorIdList {
			anchorIdList2[index] = strconv.Itoa(anchorIdList[index].(int))
		}
		sql += fmt.Sprintf(" and anchor not in (%s)", strings.Join(anchorIdList2, ","))
	}
	*/
	log.Println(sql)
	dbRows, err := dbUsersData.Db.Query(sql)
	defer dbRows.Close()
	CheckErr(err)
	uniqueUIDSet := set.NewNonTS()
	var uid, anchor int
	for dbRows.Next() {
		dbRows.Scan(&uid, &anchor)
		// followService.excludeUIDSet.Add(uid) // record
		// followService.excludeAnchorIdSet.Add(anchor)
		uniqueUIDSet.Add(uid)
		uniqueUIDSet.Add(anchor)
	}
	uidChan := make(chan int, 20) // 10
	for {
		puid := uniqueUIDSet.Pop() // 14
		if puid == nil {
			break
		}
		opuid := puid.(int)
		uidChan <- opuid
		go followService.CalculateUIDFollowFansCnt(opuid, uidChan)
	}
	emptyChanCnt := cap(uidChan)
	for i := 0; i < emptyChanCnt ; i++  {
		uidChan <- i
	}
}

// CalculateUIDFollowFansCnt 计算单个UID的粉丝数, 关注数, 存放到集合中
func (followService *FollowService) CalculateUIDFollowFansCnt(uid int, uidChan chan int)  {
	dbUsersData, err := GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
	CheckErr(err)

	followCntSet := set.New()
	fansCntSet := set.New()
	friendsCntSet := set.New()
	var followSql, fansSql, tablename string
	var anchor , friendsCnt int
	for index := 0; index < 10 ; index++  {
		tablename = USER_FOLLOW_TABLE_PREFIX + strconv.Itoa(index)
		followSql = fmt.Sprintf("select anchor from %s where uid = %d and isFriends = 0 and status = 1", tablename, uid)
		//log.Println(followSql)
		followRows, err := dbUsersData.Db.Query(followSql)
		CheckErr(err)
		for followRows.Next() {
			followRows.Scan(&anchor)
			followCntSet.Add(anchor)
		}
		followRows.Close()

		fansSql = fmt.Sprintf("select uid from %s where anchor = %d and isFriends = 0 and status = 1", tablename, uid)
		//log.Println(fansSql)
		fansRows, err := dbUsersData.Db.Query(fansSql)
		CheckErr(err)
		for fansRows.Next() {
			fansRows.Scan(&anchor)
			fansCntSet.Add(anchor)
		}
		fansRows.Close()

		friendsSql := fmt.Sprintf("select uid from %s　where uid = %d and isFriends = 1 and status = 1", tablename, uid)
		friendsRows, err := dbUsersData.Db.Query(friendsSql)
		CheckErr(err)

		for friendsRows.Next() {
			friendsRows.Scan(&friendsCnt)
			friendsCntSet.Add(friendsCnt)
		}
		friendsRows.Close()
	}

	peerUID := PeerUID{UID:uid, FollowCnt:followCntSet, FansCnt:fansCntSet, FriendsCnt: friendsCntSet}
	followService.Traffic <- peerUID
	<- uidChan
}

