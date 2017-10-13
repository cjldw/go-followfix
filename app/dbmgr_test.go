package app

import (
	"testing"
	"errors"
)

func TestNewDbMgr(t *testing.T) {
	conf := NewAppConf()
	conf.Load("../conf/app.toml", true)
	dbmgr := NewDbMgr()
	dbmgr.InitializeDbList(conf.DbConf)
	dbClient, _ := dbmgr.GetDbByName("users_data")
	t.Log(dbClient)
}

func TestSeelog(t *testing.T)  {
	err := errors.New("hahahahahh")
	CheckErr(err)
}
