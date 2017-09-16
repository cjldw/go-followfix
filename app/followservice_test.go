package app

import (
	"testing"
	"sync"
)

func TestFollowService_Produce(t *testing.T) {

	//dbUsersData, err := GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
	//CheckErr(err)
	//defer dbUsersData.Db.Close()
	//t.Log(dbUsersData.Db)
	//dbUsersData, _ = GetApp().dbmgr.GetDbByName(DB_USERS_DATA)
	//t.Log(dbUsersData.Db)
	GetApp()
	followService := NewFollowService()
	wg := &sync.WaitGroup{}
	RunAsync(wg, followService.Produce)
	RunAsync(wg, followService.Consumer)
	wg.Wait()
}

func TestChan(t *testing.T)  {

	c := make(chan int,7)
	for index := 0; index < 5 ; index++  {
		c <- 1
		go (func(c chan int, index int) {
			t.Logf("GO ROUTINE %d", index)
			<- c
		})(c, index)
	}

	for i := 0; i < 7 ; i++  {
		c <- 1
	}
}
