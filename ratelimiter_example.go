package ratelimiter_example

import (
	"fmt"
	"github.com/Mycunycu/ratelimiter"
	"time"
)

func main() {
	taskChan := make(chan func())

	limiter := ratelimiter.New(taskChan, 2, 10, time.Minute)
	limiter.DoWork()

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		counter := 0

		for range ticker.C {
			if counter == 13 {
				ticker.Stop()
				close(taskChan)
				return
			}

			task := func() {
				counter++
				count := counter

				time.Sleep(time.Second)
				fmt.Println("Task result", count)
			}

			taskChan <- task
		}
	}()

	limiter.Wait()
}
