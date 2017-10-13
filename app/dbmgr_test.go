package app

import (
	"testing"
	"errors"
	"fmt"
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


func TestQuery(t *testing.T)  {
	conf := NewAppConf()
	conf.Load("../conf/app.toml", true)
	t.Log(conf)
	dbmgr := NewDbMgr()
	dbmgr.InitializeDbList(conf.DbConf)
	pool := make(chan int, 100)
	for  {
		pool <- 10

		db, err := dbmgr.GetDbByName(DB_USERS_DATA)
		CheckErr(err)
		rows, err := db.Db.Query("select id, uid from user_follow limit 10")
		defer rows.Close()
		CheckErr(err)

		var (
			id int
			uid int
		)
		for rows.Next() {
			rows.Scan(&id)
			rows.Scan(&uid)
			fmt.Println(fmt.Sprintf("UID【%d】 ID【%d】"), id, uid)
		}
	}
}
