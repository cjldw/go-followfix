package app

import (
	"testing"
	"errors"
)

func TestNewDbMgr(t *testing.T) {
	GetApp()
	dbmgr := NewDbMgr()
	dbClient, _ := dbmgr.GetDbByName("users_data")
	t.Log(dbClient)
}

func TestSeelog(t *testing.T)  {
	err := errors.New("hahahahahh")
	CheckErr(err)
}
