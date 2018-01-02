package app

import (
	"fmt"
	"strconv"
	"sync"
	"github.com/go-redis/redis"
	log "github.com/cihub/seelog"
	"time"
	"strings"
)

var followService *FollowService

type FollowService struct {
	excludeUIDSet      map[int]int
	excludeAnchorIdSet map[int]int
	Lock               *sync.RWMutex
	Traffic            chan PeerUID
}

type PeerUID struct {
	UID        int
	FollowCntSet  map[int]int
	FansCntSet    map[int]int
	FriendsCntSet map[int]int
}

var followOnce *sync.Once = &sync.Once{}

func NewFollowService() *FollowService {
	if followService == nil {
		followOnce.Do(func() {
			followService = &FollowService{
				excludeUIDSet: make(map[int]int),
				excludeAnchorIdSet: make(map[int]int),
				Lock: &sync.RWMutex{},
				Traffic: make(chan PeerUID, 1000),
			}
		})
	}
	return followService
}

func (followService *FollowService) Produce() {
	wg := &sync.WaitGroup{}
	for table := 0; table < USER_FOLLOW_SPLIT_TABLE_NUM; table++ {
		tableName := USER_FOLLOW_TABLE_PREFIX + strconv.Itoa(table)
		wg.Add(1)
		go (func() {
			followService.processSplitTable(tableName)
			wg.Done()
		})()
	}
	wg.Wait()
	close(followService.Traffic)
	log.Info("--------所有用户PeeUID构造完毕------------")
}

func (f *FollowService) ProduceOnlyHalfMouth()  {
	defer close(f.Traffic)
	dbContents, err := GetApp().dbmgr.GetDbByName(DB_CONTENTS)
	if err != nil {
		log.Errorf("get db connection error %v", err)
	}
	sql := fmt.Sprintf("select uid from rooms where editTime >= '%s'", time.Now().AddDate(0, 0, -15).Format("2006-01-02 15:04:05"))
	log.Infof("查询半个月活动的主播: %s", sql)
	rows, err := dbContents.Query(sql)
	defer rows.Close()
	if err != nil {
		log.Critical(err)
	}
	uniqueUIDSet := make(map[int]int)
	for rows.Next() {
		var uid int
		rows.Scan(&uid)
		uniqueUIDSet[uid] = uid
	}
	log.Infof("处理半个月有活动的主播UID: %v\n", uniqueUIDSet)
	if len(uniqueUIDSet) == 0 {
		return
	}
	f.appendTraficChan(uniqueUIDSet)
}

// ProcessLoginUID
func (f *FollowService) ProcessLoginUID()  {
	defer close(f.Traffic)
	dbLog, err := GetApp().dbmgr.GetDbByName(DB_LOG)
	if err != nil {
		log.Errorf("连接数据库%s失败 :%v", DB_LOG, err)
	}

	var loopNum int = 30000
	for minId := 480001; minId < 21828590 ; minId += loopNum { // 每次处理1w条数据
		var uniqueUIDSet map[int]int = make(map[int]int)
		loginSql := fmt.Sprintf("select uid from login_log_1712 where id >= %d and id < %d group by uid",
			minId, minId + loopNum)
		rows, err := dbLog.Query(loginSql)
		if err != nil {
			log.Infof("查询数据失败: %s 错误: %v", loginSql, err)
			continue
		}
		for rows.Next() {
			var uid int
			rows.Scan(&uid)
			if uid > 0 {
				uniqueUIDSet[uid] = uid
			}

		}
		log.Infof("处理用户登录数据: %d", len(uniqueUIDSet))
		if len(uniqueUIDSet) > 0 {
			f.appendTraficChan(uniqueUIDSet)
		}
	}
}

func (f *FollowService) ProduceUIDList()  {
	defer close(f.Traffic)
	UIDConfString := GetApp().confmgr.HotfixUIDList
	UIDList := strings.Split(UIDConfString, ",")
	log.Infof("读取配置文件UID: %v", UIDList)

	if len(UIDList) == 0 {
		log.Infof("配置文件UID有误")
	}

	uniqueUIDSet := make(map[int]int)
	for _, UID := range UIDList {
		intUID, err := strconv.Atoi(UID)
		if err != nil {
			log.Errorf("配置文件UID: %v 配置异常", UID)
			continue
		}
		log.Info(intUID)
		uniqueUIDSet[intUID] = intUID
	}
	log.Infof("添加到 TrafficChan 长:%d  实体:%v ", len(uniqueUIDSet), uniqueUIDSet)
	if len(uniqueUIDSet) > 0 {
		log.Infof("添加到TrafficChan22: %v ", uniqueUIDSet)
		f.appendTraficChan(uniqueUIDSet)
	}
}

func (f *FollowService) ProcessDirtyData()  {
	defer close(f.Traffic)
	wg := &sync.WaitGroup{}
	for table := 0; table < USER_FOLLOW_SPLIT_TABLE_NUM; table++ {
		tableName := USER_FOLLOW_TABLE_PREFIX + strconv.Itoa(table)
		log.Info(tableName)
		wg.Add(1)
		go (func() {
			f.deleteDirtyData(tableName)
			wg.Done()
		})()
	}
	wg.Wait()
	log.Info("删除自己关注自己, 自己粉丝自己完成")
}

func (f *FollowService) deleteDirtyData(tableName string)  {
	dbContents, err := GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
	CheckErr(err)
	const PROCESS_MAX_ROW  = 100000;
	maxIdSql := fmt.Sprintf("select max(id) from %s ", tableName)
	log.Info(maxIdSql)
	var maxId int
	dbContents.QueryRow(maxIdSql).Scan(&maxId)
	log.Infof("max id %d", maxId)

	socialRedis, err := GetApp().redismgr.GetRedisByName(REDIS_SOCIAL)
	CheckErr(err)

	for offset := PROCESS_MAX_ROW; offset <= maxId; offset += PROCESS_MAX_ROW  {
		UIDSql := fmt.Sprintf("select uid from %s where id <= %d group by uid", tableName, offset)
		log.Info(UIDSql)
		UIDResult, err := dbContents.Query(UIDSql)
		CheckErr(err)
		for UIDResult.Next() {
			var uId int
			UIDResult.Scan(&uId)
			if uId == 0 {
				continue
			}

			friendKey := fmt.Sprintf("%s%d", FRIEND_SYSTEM_USER_FRIENDS, uId)
			followKey := fmt.Sprintf("%s%d", FRIEND_SYSTEM_USER_FOLLOW, uId)
			fansKey := fmt.Sprintf("%s%d", FRIEND_SYSTEM_USER_FANS, uId)

			socialRedis.ZRem(friendKey, uId)
			socialRedis.ZRem(followKey, uId)
			socialRedis.ZRem(fansKey, uId)
		}
	}

}

func (followService *FollowService) Consumer() {

	/* for peerUID := range followService.Traffic {
		followService.WriteDbRedis(peerUID)
	} */

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(PROCESS_UID_VAVEL)
	for i := 0; i < PROCESS_UID_VAVEL; i++ {
		go func() {
			//lock := &sync.RWMutex{}
			for peerUID := range followService.Traffic {
				followService.WriteDbRedis(peerUID)
			}
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	log.Info("--------所有用户PeeUID消费完毕------------")
}

// WriteDbRedis 将单个UID用户写入到Reids中, 更新数据库
func (followService *FollowService) WriteDbRedis(peerUID PeerUID) {
	uId := peerUID.UID
	redisSocial, err := GetApp().redismgr.GetRedisByName(REDIS_SOCIAL)
	CheckErr(err)
	var followListKey string = fmt.Sprintf("%s%d", FRIEND_SYSTEM_USER_FOLLOW, uId)
	var fansListKey string = fmt.Sprintf("%s%d", FRIEND_SYSTEM_USER_FANS, uId)
	var friendsListKey string = fmt.Sprintf("%s%d", FRIEND_SYSTEM_USER_FRIENDS, uId)

	var oldFollowList string = fmt.Sprintf(USER_FOLLOW_LIST, uId)
	// Fetch UID's Follow List And Storage To Social Redis
	//fmt.Printf("UID[%d] 关注数量: %d 关注列表: %v \n", uId, len(peerUID.FollowCntSet), peerUID.FollowCntSet)
	/* if len(peerUID.FollowCntSet) > 0 {
		WriteLog("d:/rediskey.log", followListKey)
	} */
	nowTime := float64(time.Now().Unix())
	for _, followUID := range peerUID.FollowCntSet {
		//WriteLog("/tmp/uid_follow.log", fmt.Sprintf("%v", iFollowUID))
		followUIDFansCnt := getUIDFansCnt(followUID)
		item := redis.Z{
			Score:  float64(followUIDFansCnt),
			Member: followUID,
		}
		err := redisSocial.ZAdd(followListKey, item).Err()
		if err != nil {
			log.Error(err)
		}
		followItem := redis.Z{
			Score: nowTime,
			Member: followUID,
		}
		// 老的关注集合处理
		err = redisSocial.ZAdd(oldFollowList, followItem).Err()
		if err != nil {
			log.Error(err)
		}
	}
	//fmt.Printf("UID[%d] 粉丝数量: %d 粉丝列表: %v \n", uId, len(peerUID.FansCntSet), peerUID.FansCntSet)
	/* if len(peerUID.FansCntSet) > 0 {
		WriteLog("d:/rediskey.log", fansListKey)
	} */
	for _, fansUID := range peerUID.FansCntSet {
		fansUIDFansCnt := getUIDFansCnt(fansUID)
		item := redis.Z{
			Score:  float64(fansUIDFansCnt),
			Member: fansUID,
		}
		err := redisSocial.ZAdd(fansListKey, item).Err()
		if err != nil {
			log.Error(err)
		}

		// 当前用户添加到关注对象的粉丝集合中
		oldFansListKey := fmt.Sprintf(USER_FANS_LIST, fansUID)
		fansItem := redis.Z{
			Score: nowTime,
			Member: uId,
		}
		err = redisSocial.ZAdd(oldFansListKey, fansItem).Err()
		if err != nil {
			log.Error(err)
		}
	}

	/* if len(peerUID.FriendsCntSet) > 0 {
		WriteLog("d:/rediskey.log", friendsListKey)
	} */
	for _, friendUID := range peerUID.FriendsCntSet {
		friendsUIDFansCnt := getUIDFansCnt(friendUID)
		item := redis.Z{
			Score:  float64(friendsUIDFansCnt),
			Member: friendUID,
		}
		err := redisSocial.ZAdd(friendsListKey, item).Err()
		if err != nil {
			log.Error(err)
		}
	}
}

// getUIDFansCnt get user's fans number
// if redis cache return direct
// or select from MySQL then save and return
func getUIDFansCnt(uid int) int {
	redisSocial, err := GetApp().redismgr.GetRedisByName(REDIS_SOCIAL)
	CheckErr(err)
	fansCntKey := fmt.Sprintf("id_%d_fanscnt", uid)
	sFansCnt, err := redisSocial.HGet(TMP_UID_FANS_NUM, fansCntKey).Result()
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
	//now := time.Now()
	for index := 0; index < USER_FOLLOW_SPLIT_TABLE_NUM; index++ {
		sql = fmt.Sprintf("select count(*) as fansCnt from user_follow_%d where anchor = %d and status = 1 "+
			"and isFriends = 0", index, uid)
		dbUserData.QueryRow(sql).Scan(&fragmentCnt)
		fansCnt += fragmentCnt
	}
	//	log.Printf("UID[%s] 粉丝数量 [%d]", fansCntKey, fansCnt)
	redisSocial.HSet(TMP_UID_FANS_NUM, fansCntKey, strconv.Itoa(fansCnt))
	//	WriteLog("/tmp/time.log", fmt.Sprintf("粉丝时间： %v", time.Since(now)))
	//fmt.Printf("UID[%d]粉丝数量: %v \n", uid, time.Since(now))
	return fansCnt
}

// CalculateUIDFollowFansCnt 计算单个UID的粉丝数, 关注数, 存放到集合中
func (followService *FollowService) CalculateUIDFollowFansCnt(inputChan <-chan int, outputChan chan <- PeerUID) {
	for uId := range inputChan {
		dbUsersData, err := GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
		CheckErr(err)
		var followSql, fansSql, tableName string
		var followAnchor, fansAnchor, friendsAnchor int
		followCntSet := make(map[int]int)
		fansCntSet := make(map[int]int)
		friendsCntSet := make(map[int]int)

		for index := 0; index < USER_FOLLOW_SPLIT_TABLE_NUM; index++ {
			tableName = USER_FOLLOW_TABLE_PREFIX + strconv.Itoa(index)
			followSql = fmt.Sprintf("select anchor from %s where uid = %d and isFriends = 0 and status = 1",
				tableName, uId)
			followRows, err := dbUsersData.Query(followSql)
			CheckErr(err)
			for followRows.Next() {
				followRows.Scan(&followAnchor)
				followCntSet[followAnchor] = followAnchor
			}
			followRows.Close()
			fansSql = fmt.Sprintf("select uid from %s where anchor = %d and isFriends = 0 and status = 1",
				tableName, uId)
			fansRows, err := dbUsersData.Query(fansSql)
			CheckErr(err)
			for fansRows.Next() {
				fansRows.Scan(&fansAnchor)
				fansCntSet[fansAnchor] = fansAnchor
			}
			fansRows.Close()

			friendsSql := fmt.Sprintf("select uid from %s where uid = %d and isFriends = 1 and status = 1",
				tableName, uId)
			friendsRows, err := dbUsersData.Query(friendsSql)
			CheckErr(err)

			for friendsRows.Next() {
				friendsRows.Scan(&friendsAnchor)
				friendsCntSet[friendsAnchor] = friendsAnchor
			}
			friendsRows.Close()
		}
		peerUID := PeerUID{UID: uId, FollowCntSet: followCntSet, FansCntSet: fansCntSet, FriendsCntSet: friendsCntSet}
		outputChan <- peerUID
	}
}

// processSplitTable 处理分表数据
func (followService *FollowService) processSplitTable(tableName string) {
	dbUsersData, err := GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
	CheckErr(err)
	sql := fmt.Sprintf("select uid, anchor from %s where  status = 1 and isFriends = 0 and isAnchor = 1", tableName)
	log.Info(sql)
	dbRows, err := dbUsersData.Query(sql)
	defer dbRows.Close()
	CheckErr(err)
	uniqueUIDSet := make(map[int]int)
	for dbRows.Next() {
		var uid, anchor int
		dbRows.Scan(&uid, &anchor)
		uniqueUIDSet[uid] = uid
		//uniqueUIDSet[anchor] = anchor
	}
	log.Infof("表 [%s] 共[%d] 个UID \n", tableName, len(uniqueUIDSet))
	followService.appendTraficChan(uniqueUIDSet)
}

func (followService *FollowService) appendTraficChan(uniqueUIDSet map[int]int)  {
	var valve int = len(uniqueUIDSet)
	if valve > PROCESS_UID_VAVEL {
		valve = PROCESS_UID_VAVEL
	}
	log.Infof("max goroutine numbers: %d", valve)
	// 限制最大goroutine数量
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(valve)
	outputChan := make(chan PeerUID, valve)
	inputChan := make(chan int, valve)
	go func() {
		for  {
			if len(uniqueUIDSet) == 0 {
				close(inputChan)
				break
			}
			for _, uId := range uniqueUIDSet {
				inputChan <- uId
				delete(uniqueUIDSet, uId)
			}
		}
	}()
	for i := 0; i < valve ; i++ {
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
