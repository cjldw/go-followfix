package app

import (
	"testing"
	"errors"
)

func TestNewDbMgr(t *testing.T) {
	appConf := InitializeConf()
	dbmgr := NewDbMgr()
	t.Log(appConf.DbConf)
	dbmgr.InitializeDbList(appConf.DbConf)
	dbClient, _ := dbmgr.GetDbByName("users_data")
	t.Log(dbClient)
	sql := "select * from user"
	dbRows,  err := dbClient.Db.Query(sql)
	defer dbRows.Close()
	CheckErr(err)
	var userList []User
	for dbRows.Next()  {
		user := User{}
		dbRows.Scan(&user.Id, &user.Name, &user.Gender, &user.Mobile)
		t.Log(user)
		userList = append(userList, user)
	}
	t.Log(userList)
}

func TestSeelog(t *testing.T)  {
	err := errors.New("hahahahahh")
	CheckErr(err)
}
