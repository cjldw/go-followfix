package app

import "sync"

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
		app = &App{}
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

