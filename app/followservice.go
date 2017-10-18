package app

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"github.com/go-redis/redis"
	set "gopkg.in/fatih/set.v0"
	//"os"
	"time"
)

var followService *FollowService

type FollowService struct {
	excludeUIDSet      *set.Set
	excludeAnchorIdSet *set.Set
	Lock               *sync.RWMutex
	Traffic            chan PeerUID
	ProduceEnd         chan bool
}

type PeerUID struct {
	UID        int
	FollowCnt  *set.Set
	FansCnt    *set.Set
	FriendsCnt *set.Set
}

var followOnce *sync.Once = &sync.Once{}

func NewFollowService() *FollowService {
	if followService == nil {
		followOnce.Do(func() {
			followService = &FollowService{set.New(), set.New(), &sync.RWMutex{}, make(chan PeerUID, 10), make(chan bool)}
		})
	}
	return followService
}

func (followService *FollowService) Produce() {
	wg := &sync.WaitGroup{}
	for table := 0; table < USER_FOLLOW_SPLIT_TABLE_NUM; table++ {
		tablename := USER_FOLLOW_TABLE_PREFIX + strconv.Itoa(table)
		wg.Add(1)
		go (func() {
			followService.processSplitTable(tablename)
			wg.Done()
		})()
	}
	wg.Wait()
	close(followService.Traffic)
	fmt.Println("------produce ok--------------")
}

func (followService *FollowService) Consumer() {
	lock := &sync.RWMutex{}
	for peerUID := range followService.Traffic { // 单进程跑redis
		followService.WriteDbRedis(peerUID, lock)
	}
	log.Println("所有用户数据处理完毕")
}

// WriteDbRedis 将单个UID用户写入到Reids中, 更新数据库
func (followService *FollowService) WriteDbRedis(peerUID PeerUID, lock *sync.RWMutex) {
	uId := peerUID.UID
	redisSocial, err := GetApp().redismgr.GetRedisByName(REDIS_SOCIAL)
	CheckErr(err)
	var followListKey string = fmt.Sprintf("%s%d", FRIEND_SYSTEM_USER_FOLLOW, uId)
	var fansListKey string = fmt.Sprintf("%s%d", FRIEND_SYSTEM_USER_FANS, uId)
	var friendsListKey string = fmt.Sprintf("%s%d", FRIEND_SYSTEM_USER_FRIENDS, uId)

	// Fetch UID's Follow List And Storage To Social Redis
	fmt.Printf("UID[%d] 关注数量: %d 关注列表: %d \n", uId, peerUID.FollowCnt.Size(), peerUID.FollowCnt)
	for { // 处理关注
		//		log.Printf("用户[%d]关注 ->［%v］\n", uid, followUID)
		if peerUID.FollowCnt.Size() == 0 {
			break
		}
		followUID := peerUID.FollowCnt.Pop()
		iFollowUID, ok := followUID.(int)
		if !ok {
			continue
		}
		//WriteLog("/tmp/uid_follow.log", fmt.Sprintf("%v", iFollowUID))
		followUIDFansCnt := getUIDFansCnt(iFollowUID)
		item := redis.Z{
			Score:  float64(followUIDFansCnt),
			Member: followUID,
		}
		redisSocial.RedisClient.ZAdd(followListKey, item)
	}
	fmt.Printf("UID[%d] 粉丝数量: %d 粉丝列表: %d \n", uId, peerUID.FansCnt.Size(), peerUID.FansCnt)
	for { // 处理粉丝
		//		log.Printf("用户[%d]粉丝 ->［%v］\n", uid, fansUID)
		if peerUID.FansCnt.Size() == 0 {
			break
		}
		fansUID := peerUID.FansCnt.Pop()
		iFansUID, ok := fansUID.(int)
		if !ok {
			continue
		}
		fansUIDFansCnt := getUIDFansCnt(iFansUID)
		item := redis.Z{
			Score:  float64(fansUIDFansCnt),
			Member: fansUID,
		}
		redisSocial.RedisClient.ZAdd(fansListKey, item)
	}

	fmt.Printf("UID[%d] 好友数量: %d 好友列表: %d \n", uId, peerUID.FriendsCnt.Size(), peerUID.FriendsCnt)
	for { // 处理好友
		//		log.Printf("用户[%d]好有 ->［%v］\n", uid, friendsUID)
		if peerUID.FriendsCnt.Size() == 0 {
			break
		}
		friendsUID := peerUID.FriendsCnt.Pop()
		iFriendUID, ok := friendsUID.(int)
		if !ok {
			continue
		}
		friendsUIDFansCnt := getUIDFansCnt(iFriendUID)
		item := redis.Z{
			Score:  float64(friendsUIDFansCnt),
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
	sFansCnt, err := redisSocial.RedisClient.HGet(TMP_UID_FANS_NUM, fansCntKey).Result()
	if err == nil {
		iFansCnt, err := strconv.Atoi(sFansCnt)
		if err != nil {
			return 0
		}
		return iFansCnt
	}
	var sql string
	dbUserData, err := GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
	CheckErr(err)
	var fansCnt, fragmentCnt int
	now := time.Now()
	for index := 0; index < USER_FOLLOW_SPLIT_TABLE_NUM; index++ {
		sql = fmt.Sprintf("select count(*) as fansCnt from user_follow_%d where anchor = %d and status = 1 "+
			"and isFriends = 0", index, uid)
		dbUserData.Db.QueryRow(sql).Scan(&fragmentCnt)
		fansCnt += fragmentCnt
	}
	//	log.Printf("UID[%s] 粉丝数量 [%d]", fansCntKey, fansCnt)
	redisSocial.RedisClient.HSet(TMP_UID_FANS_NUM, fansCntKey, strconv.Itoa(fansCnt))
	//	WriteLog("/tmp/time.log", fmt.Sprintf("粉丝时间： %v", time.Since(now)))
	fmt.Printf("UID[%d]粉丝数量: %v \n", uid, time.Since(now))
	return fansCnt
}

// processSplitTable 处理分表数据
func (followService *FollowService) processSplitTable(tablename string) {
	dbUsersData, err := GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
	CheckErr(err)
	sql := fmt.Sprintf("select uid, anchor from %s where  status = 1 and isFriends = 0 and isAnchor = 1", tablename)
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
	fmt.Println(sql)
	dbRows, err := dbUsersData.Db.Query(sql)
	defer dbRows.Close()
	CheckErr(err)
	uniqueUIDSet := set.New()
	for dbRows.Next() {
		var uid, anchor int
		dbRows.Scan(&uid, &anchor)
		// followService.excludeUIDSet.Add(uid) // record
		// followService.excludeAnchorIdSet.Add(anchor)
		//WriteLog("/tmp/produceUID", fmt.Sprintf("%d, %d", uid, anchor))
		uniqueUIDSet.Add(uid)
		uniqueUIDSet.Add(anchor)
	}
	fmt.Printf("表 [%s] 共[%d] 个UID \n", tablename, uniqueUIDSet.Size())

	// 限制最大goroutine数量
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(PROCESS_UID_VAVEL)
	outputChan := make(chan PeerUID, PROCESS_UID_VAVEL)
	inputChan := make(chan int, PROCESS_UID_VAVEL)
	go func() {
		for  {
			if uniqueUIDSet.Size() == 0 {
				close(inputChan)
				break
			}
			uId, ok := uniqueUIDSet.Pop().(int)
			if !ok {
				continue
			}
			inputChan <- uId
		}
	}()
	for i := 0; i < PROCESS_UID_VAVEL ; i++ {
		go func() {
			followService.CalculateUIDFollowFansCnt(inputChan, outputChan)
			waitGroup.Done()
		}()
	}

	go func() {
		waitGroup.Wait()
		close(outputChan)
	}()

	for peerUID := range outputChan {
		followService.Traffic <- peerUID
	}
}

// CalculateUIDFollowFansCnt 计算单个UID的粉丝数, 关注数, 存放到集合中
func (followService *FollowService) CalculateUIDFollowFansCnt(inputChan <-chan int, outputChan chan <- PeerUID) {
	for uId := range inputChan{
		dbUsersData, err := GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
		CheckErr(err)
		var followSql, fansSql, tablename string
		var followAnchor, fansAnchor, friendsCnt int
		followCntSet := set.New()
		fansCntSet := set.New()
		friendsCntSet := set.New()

		for index := 0; index < USER_FOLLOW_SPLIT_TABLE_NUM; index++ {
			tablename = USER_FOLLOW_TABLE_PREFIX + strconv.Itoa(index)
			followSql = fmt.Sprintf("select anchor from %s where uid = %d and isFriends = 0 and status = 1", tablename, uId)
			//log.Println(followSql)
			followRows, err := dbUsersData.Db.Query(followSql)
			CheckErr(err)
			for followRows.Next() {
				followRows.Scan(&followAnchor)
				followCntSet.Add(followAnchor)
				//WriteLog("/tmp/produce_follow_uid.log", fmt.Sprintf("%v", followAnchor))
			}
			followRows.Close()

			fansSql = fmt.Sprintf("select uid from %s where anchor = %d and isFriends = 0 and status = 1", tablename, uId)
			//log.Println(fansSql)
			fansRows, err := dbUsersData.Db.Query(fansSql)
			CheckErr(err)
			for fansRows.Next() {
				fansRows.Scan(&fansAnchor)
				fansCntSet.Add(fansAnchor)
			}
			fansRows.Close()

			friendsSql := fmt.Sprintf("select uid from %s where uid = %d and isFriends = 1 and status = 1", tablename, uId)
			friendsRows, err := dbUsersData.Db.Query(friendsSql)
			CheckErr(err)

			for friendsRows.Next() {
				friendsRows.Scan(&friendsCnt)
				friendsCntSet.Add(friendsCnt)
			}
			friendsRows.Close()
		}
		peerUID := PeerUID{UID: uId, FollowCnt: followCntSet, FansCnt: fansCntSet, FriendsCnt: friendsCntSet}
		outputChan <- peerUID
		fmt.Println(peerUID)
	}
}
