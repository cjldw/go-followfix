package app

import (
	"sync"
	log "github.com/cihub/seelog"
)

type App struct {
	confmgr *AppConf
	dbmgr *DbMgr
	redismgr *RedisMgr
	isInitialize bool
}

var app *App
var appOnce *sync.Once = &sync.Once{}

func GetApp() *App  {
	if app == nil {
		app = NewApp()
	}
	return app
}

// NewApp return singleton application instance
func NewApp() *App  {
	if app != nil && app.isInitialize {
		return app
	}
	appOnce.Do(func() {
		app = &App{isInitialize:true}
		app.confmgr = (func() *AppConf{
			appconf, err := NewAppConf().Load(APP_CONFIG_PATH, false)
			CheckErr(err)
			return appconf
		})()
		app.dbmgr = (func() *DbMgr{
			dbmgr := NewDbMgr()
			dbmgr.InitializeDbList(app.confmgr.DbConf)
			return dbmgr

		})()
		app.redismgr = (func() *RedisMgr{
			redismgr := NewRedisMgr()
			redismgr.InitializeRedisList(app.confmgr.RedisConf)
			return redismgr
		})()
		InitLog("conf/log4go.xml")
	})
	return app
}

// Run Application Entrance
func (app *App) Run()  {
	wg := &sync.WaitGroup{}
	followService := NewFollowService()
	switch app.confmgr.HotfixType {
	case "mounth":
		log.Info("处理半个月活动UID")
		RunAsync(wg, followService.ProduceOnlyHalfMouth)
	case "uid":
		log.Infof("处理UID: %s", app.confmgr.HotfixUIDList)
		if app.confmgr.HotfixUIDList != "" {
			RunAsync(wg, followService.ProduceUIDList)
		}
	case "self":
		log.Info("处理自己关注自己UID")
		RunAsync(wg, followService.ProcessDirtyData)
	case "login":
		log.Info("处理半个月登入的用户")
		RunAsync(wg, followService.ProcessLoginUID)
	default:
		log.Info("默认处理所有数据")
		RunAsync(wg, followService.Produce)
	}
	RunAsync(wg, followService.Consumer)
	wg.Wait()
}

