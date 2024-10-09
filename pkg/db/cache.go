package db

import (
	"fmt"
	"os"
	"path"
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

	userDataDir := path.Join(dataDir, userID)
	if err := ensureDirExists(userDataDir); err != nil {
		return nil, err
	}

	db, err := NewStore(userDataDir, contentTypes)
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

func ensureDirExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	} else if err != nil {
		// 其他错误
		return fmt.Errorf("failed to check directory: %w", err)
	}
	return nil
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
