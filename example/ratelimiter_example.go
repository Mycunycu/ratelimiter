package example

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Mycunycu/ratelimiter"
)

func Example() {
	taskChan := make(chan func())

	limiter := ratelimiter.New(taskChan, 2, 5, time.Minute)
	limiter.DoWork()

	// Task provider simulating
	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		var counter int32

		for range ticker.C {
			// as example closing the task channel
			if atomic.LoadInt32(&counter) == 13 {
				ticker.Stop()
				close(taskChan)
				return
			}

			atomic.AddInt32(&counter, 1)
			task := func() {
				count := atomic.LoadInt32(&counter)

				time.Sleep(time.Second)
				fmt.Println("Task result", count)
			}

			taskChan <- task
		}
	}()

	limiter.Wait()
}

// Task result 1
// Task result 2
// Task result 4
// Task result 3
// Task result 5

//               <- waiting for update limits in the time window

// Task result 7
// Task result 6
// Task result 9
// Task result 8
// Task result 10

//               <- waiting for update limits in the time window

// Task result 12
// Task result 11
// Task result 13
// Task channel are closed
