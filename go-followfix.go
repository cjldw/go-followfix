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
	startProfile := time.Now()
	app := app.NewApp()
	app.Run()
	endProfile := time.Now()
	log.Printf("总共花费 【%f.2】秒！", endProfile.Sub(startProfile).Seconds())

}
