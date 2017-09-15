package app

import "testing"

func TestFollowService_Produce(t *testing.T) {

	//dbUsersData, err := GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
	//CheckErr(err)
	//defer dbUsersData.Db.Close()
	//t.Log(dbUsersData.Db)
	//dbUsersData, _ = GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
	//t.Log(dbUsersData.Db)
	GetApp()
	followService := NewFollowService()
	followService.Produce()
}
