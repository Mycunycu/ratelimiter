package ratelimiter

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// RateLimiter object
type RateLimiter struct {
	chanIn        chan func()
	rateChan      chan struct{}
	maxTask       uint32
	limit         uint32
	limitInterval time.Duration
	count         uint32
	wg            *sync.WaitGroup
}

// New - RateLimiter constructor
// taskChan - channel for getting a tasks
// max - maximum allowed count parallel tasks
// limit - rate limit in a time window
// interval - time window with limit
func New(taskChan chan func(), max, limit uint32, interval time.Duration) *RateLimiter {
	if max == 0 {
		max = 1
	}

	if limit == 0 {
		limit = 50
	}

	return &RateLimiter{
		chanIn:        taskChan,
		rateChan:      make(chan struct{}, limit),
		maxTask:       max,
		limit:         limit,
		limitInterval: interval,
		wg:            &sync.WaitGroup{},
	}
}

// DoWork - run to working the instance RateLimiter
func (rl *RateLimiter) DoWork() {
	rl.initLimit()

	for i := 0; i < int(rl.maxTask); i++ {
		rl.wg.Add(1)
		go rl.taskWorker()
	}

	go rl.renewLimitByInterval()
}

func (rl *RateLimiter) initLimit() {
	for i := 0; i < int(rl.limit); i++ {
		rl.rateChan <- struct{}{}
	}
}

func (rl *RateLimiter) taskWorker() {
	for {
		select {
		case task, ok := <-rl.chanIn:
			if !ok {
				rl.wg.Done()
				return
			}

			<-rl.rateChan
			atomic.AddUint32(&rl.count, 1)
			task()
		}
	}

}

func (rl *RateLimiter) renewLimitByInterval() {
	ticker := time.NewTicker(rl.limitInterval)

	for range ticker.C {
		rl.renewLimit()
	}
}

func (rl *RateLimiter) renewLimit() {
	needToAdd := atomic.LoadUint32(&rl.count)
	atomic.StoreUint32(&rl.count, 0)

	for i := 0; i < int(needToAdd); i++ {
		rl.rateChan <- struct{}{}
	}
}

// Wait - waiting while all workers which process tasks in work
// The signal to stop working - close the task chanel
// IMPORTANT - close the channel with caution to avoid panic, don't write into the closed channel.
func (rl *RateLimiter) Wait() {
	rl.wg.Wait()
	fmt.Println("Task chan is closed")
}
