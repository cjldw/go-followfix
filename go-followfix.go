package main

import (
	"github.com/vvotm/go-followfix/app"
	"runtime"
	"time"
	"log"
)

func main()  {
	numCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)
	prof := time.Now()
	appInst := app.NewApp()
	go TimerUpdate()
	appInst.Run()
	log.Printf("总共花费 【%f.2】秒！", time.Since(prof).Seconds())


}


func TimerUpdate() {
	timer1 := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-timer1.C:
			//打印统计信息
			//postesp, postcnt := PostSvrMgrIns().GetAverage()
			//log4go.Info("Average post:%v--%v", postesp, postcnt)
			log.Printf("goroutine数量 %d\n", runtime.NumGoroutine())
		}
	}
}
