package app

import "sync"

type App struct {
	confmgr *AppConf
	dbmgr *DbMgr
	redismgr *RedisMgr
	isInitialize bool
}

var app *App
var once *sync.Once = sync.Once{}

func GetApp() *App  {
	if app == nil {
		return NewApp()
	}
	return app
}

func NewApp() *App  {
	if app != nil && app.isInitialize {
		return app
	}
	once.Do(func() {
		app = &App{}
		app.confmgr = (func() {
			appconf, err := NewAppConf().Load(APP_CONFIG_PATH, false)
			CheckErr(err)
			return appconf
		})()
		app.dbmgr = (func() {
			dbmgr := NewDbMgr()
			dbmgr.InitializeDbList(app.confmgr.DbConf)
			return dbmgr

		})()
		app.redismgr = (func() {
			redismgr := NewRedisMgr()
			redismgr.InitializeRedisList(app.confmgr.RedisConf)
			return redismgr
		})()
	})
	return app
}

// Run Application Entrance
func (app *App) Run()  {
	wg := &sync.WaitGroup{}
	followService := NewFollowService()
	RunAsync(wg, followService.Produce)
	RunAsync(wg, followService.Consumer)
	wg.Wait()
}

