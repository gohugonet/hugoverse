package db

import "sync"

var mu = &sync.Mutex{}
var configCache map[string]interface{}

// ConfigCache is a in-memory cache of the Configs for quicker lookups
// 'key' is the JSON tag associated with the config field
func ConfigCache(key string) interface{} {
	mu.Lock()
	val := configCache[key]
	mu.Unlock()

	return val
}
