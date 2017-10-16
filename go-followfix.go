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
	appInst.Run()
	log.Printf("总共花费 【%f.2】秒！", time.Since(prof).Seconds())

}
