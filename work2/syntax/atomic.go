package syntax

import (
	"sync"
	"sync/atomic"
)

func main() {
	var val int64 = 0
	var wg sync.WaitGroup

	for i := 0; i <= 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			valp := atomic.AddInt64(&val, 1)
			println(valp)
		}()
	}
	wg.Wait()

}
