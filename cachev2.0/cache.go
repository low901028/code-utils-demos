package cache_go

import (
	"sync"
)

var (
	cache = make(map[string]*CacheTable)  // 存放存储空间，name唯一： name<->cacheTable
	mutex sync.RWMutex
)

// 若是已存在cache则直接返回 或者新建并返回
func Cache(table string) *CacheTable {
	mutex.RLock()
	t, ok := cache[table]
	mutex.RUnlock()

	if !ok {
		mutex.Lock()
		t, ok = cache[table]
		// Double check whether the table exists or not.
		if !ok {
			t = &CacheTable{
				name:  table,
				items: make(map[interface{}]*CacheItem),
			}
			cache[table] = t
		}
		mutex.Unlock()
	}

	return t
}
