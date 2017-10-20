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
	db, err := dbmgr.GetDbByName(DB_USERS_DATA)
	CheckErr(err)
	rows, err := db.Query("select id, uid from user_follow")
	defer rows.Close()

	for rows.Next() {
		var (
			id int
			uid int
		)
		rows.Scan(&id)
		rows.Scan(&uid)

		t.Log(id, uid)
	}

	return

	return;
	pool := make(chan int, 100)
	for  {
		pool <- 10
		db, err := dbmgr.GetDbByName(DB_CONTENTS)
		CheckErr(err)
		rows, err := db.Query("select id, name from games")
		CheckErr(err)
		var (
			id int
			name string
		)
		for rows.Next() {
			rows.Scan(&id)
			rows.Scan(&name)
			fmt.Println(fmt.Sprintf("UID【%d】 ID【%s】", id, name));
		}
		rows.Close()
	}
}
