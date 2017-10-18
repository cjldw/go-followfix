package app

import (
	"sync"
	"os"
	"bufio"
)

// RunAsync
func RunAsync(wg *sync.WaitGroup, fn func())  {
	wg.Add(1)
	go (func() {
		fn()
		wg.Done()
	})()
}

func WriteLog(dir, line string) {
	f, err := os.OpenFile(dir, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f) //创建新的 Writer 对象
	//t := time.Now().Format("2006-01-02 15:04:05")
	_, err = w.WriteString(/*t + ":"+ */ line + "\n")
	if err != nil {
		panic(err)
	}

	w.Flush()
}

