package main

import (
	"github.com/vvotm/go-followfix/app"
	"runtime"
	"time"
	log "github.com/cihub/seelog"
	"fmt"
)

func main()  {
	numCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)
	prof := time.Now()
	appInst := app.NewApp()
	log.Info("------------------------------------------")
	// go TimerUpdate()
	appInst.Run()
	fmt.Println(fmt.Sprintf("总共花费 【%f.2】秒！", time.Since(prof).Seconds()))


}


func TimerUpdate() {
	timer1 := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-timer1.C:
			//打印统计信息
			//postesp, postcnt := PostSvrMgrIns().GetAverage()
			//log4go.Info("Average post:%v--%v", postesp, postcnt)
			log.Infof("goroutine数量 %d", runtime.NumGoroutine())
		}
	}
}
