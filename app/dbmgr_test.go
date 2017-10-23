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
	rows, err := db.Query("select a, b from tx_test limit 10")
	defer rows.Close()
	CheckErr(err)
	for rows.Next() {
		var (
			a int
			b int
		)
		rows.Scan(&a, &b)
		fmt.Println(fmt.Sprintf("UID【%d】 ID【%d】", a, b))
	}
}
