package main

import (
	"github.com/vvotm/go-followfix/app"
	"runtime"
)

func main()  {
	numCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)
	app := app.NewApp()
	app.Run()
}
