package app

import (
	"testing"
	"sync"
	"github.com/go-redis/redis"
	"errors"
	"fmt"
	"time"
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

func TestHashGet(t *testing.T)  {
	var a interface{} = 22
	t.Log(a)
	aa, ok := a.(int)
	if !ok {
		t.Fatal(errors.New("convert interface{} to int error !"))
	}
	t.Log(aa + 1)

	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		Password: "",
		DB: 0,
	})
	val, err := redisClient.HGet("user", "age").Result()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(val + "a")

}

func TestGetUIDFansCnt(t *testing.T)  {

	demo := make(chan int, 10)
	wg := &sync.WaitGroup{}
	go func() {
		for item := range demo {
			wg.Add(1)
			time.Sleep(2 * time.Second)
			fmt.Println(item)
			wg.Done()

		}
	}()
	for i := 0 ; i < 10 ; i++  {
		demo <- i
	}

	wg.Wait()

	t.Log("-----------------")
}
