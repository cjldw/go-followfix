package app

import "sync"

// RunAsync
func RunAsync(wg *sync.WaitGroup, fn func())  {
	wg.Add(1)
	go (func() {
		fn()
		wg.Done()
	})()
}
