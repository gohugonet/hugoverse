package db

import (
	"sync"
	"time"
)

type cachedStore struct {
	*Store

	timer *time.Timer
}

func (cs *cachedStore) close() error {
	cs.timer.Stop()
	return cs.Store.Close()
}

var (
	cache               = make(map[string]*cachedStore)
	mu                  sync.Mutex
	cleanupWaitDuration = 10 * time.Minute
)

func OpenUserStore(userID string, dataDir string, contentTypes []string) (*Store, error) {
	mu.Lock()
	defer mu.Unlock()

	if cachedDB, ok := cache[userID]; ok {
		resetDBTimer(cachedDB)
		return cachedDB.Store, nil
	}

	db, err := NewStore(dataDir, contentTypes)
	if err != nil {
		return nil, err
	}

	cs := &cachedStore{
		Store: db,
	}
	resetDBTimer(cs)

	cache[userID] = cs
	return cs.Store, nil
}

func resetDBTimer(cachedDB *cachedStore) {
	if cachedDB.timer != nil {
		cachedDB.timer.Stop()
	}

	cachedDB.timer = time.NewTimer(cleanupWaitDuration)
	go func() {
		<-cachedDB.timer.C
		cleanupIdleDB(cachedDB) // 触发统一的清理函数
	}()
}

func cleanupIdleDB(idleDB *cachedStore) {
	mu.Lock()
	defer mu.Unlock()

	for userID, db := range cache {
		if db == idleDB {
			err := db.close()
			if err != nil {
				db.log.Errorln("Couldn't close db.", err)
				return
			}

			delete(cache, userID)
			return
		}
	}
}
