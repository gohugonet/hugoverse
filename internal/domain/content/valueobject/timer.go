package valueobject

import (
	"log"
	"sync"
	"time"
)

var sortContentCalls = make(map[string]time.Time)
var waitDuration = time.Millisecond * 2000
var sortMutex = &sync.Mutex{}

func setLastInvocation(key string) {
	sortMutex.Lock()
	sortContentCalls[key] = time.Now()
	sortMutex.Unlock()
}

func lastInvocation(key string) (time.Time, bool) {
	sortMutex.Lock()
	last, ok := sortContentCalls[key]
	sortMutex.Unlock()
	return last, ok
}

func EnoughTime(key string, cb func(key string) error) bool {
	last, ok := lastInvocation(key)
	if !ok {
		// no invocation yet
		// track next invocation
		setLastInvocation(key)
		return true
	}

	// if our required wait time has been met, return true
	if time.Now().After(last.Add(waitDuration)) {
		setLastInvocation(key)
		return true
	}

	// dispatch a delayed invocation in case no additional one follows
	go func() {
		lastInvocationBeforeTimer, _ := lastInvocation(key) // zero value can be handled, no need for ok
		enoughTimer := time.NewTimer(waitDuration)
		<-enoughTimer.C
		lastInvocationAfterTimer, _ := lastInvocation(key)
		if !lastInvocationAfterTimer.After(lastInvocationBeforeTimer) {
			log.Println("Time to trigger sort", key)
			if err := cb(key); err != nil {
				log.Println("Error while updating db with sorted", key, err)
				return
			}
		}
	}()

	return false
}
