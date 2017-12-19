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
	dbmgr := NewDbMgr()
	dbmgr.InitializeDbList(conf.DbConf)
	db, err := dbmgr.GetDbByName(DB_CONTENTS)
	CheckErr(err)
	var uid int
	var title string
	db.QueryRow("select max(uid) as maxId, title from rooms limit 1").Scan(&uid, &title)
	t.Log(uid, title)

	return
	rows, err := db.Query("select uid from rooms limit 1")
	defer rows.Close()

	for rows.Next() {
		var (
			uid int
		)
		rows.Scan(&uid)

		t.Log(uid)
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
