package cache_go

import (
	"sync"
	"time"
)

// cache中的item
type CacheItem struct {
	sync.RWMutex

	key interface{} 		// item-key
	data interface{} 		// item-value
	lifeSpan time.Duration 	// item exipre period
	createOn time.Time      // item create time
	accessedOn time.Time    // item last access timestamp
	accessCount int64       // item access count

	aboutToExpire 	func(key interface{}) // remove the item from the cache： callback method
}

// 新建item
func NewCacheItem(key interface{}, lifeSpan time.Duration, data interface{})  *CacheItem {
	t := time.Now()
	return &CacheItem{
		key:		key,
		lifeSpan: lifeSpan,
		createOn:	t,
		accessedOn: t,
		accessCount: 0,
		aboutToExpire: nil,
		data: data,
	}
}

// 保持item有效
func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedOn = time.Now()
	item.accessCount++
}

// 只读
func (item *CacheItem) LifeSpan() time.Duration {
	// immutable
	return item.lifeSpan
}

// 防止多个程序访问同一个item：使用RLock
func (item *CacheItem) AccessedOn() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessedOn
}

//
func (item *CacheItem) CreatedOn() time.Time {
	// immutable
	return item.createOn
}

// 防止多个程序访问同一个item：使用RLock
func (item *CacheItem) AccessCount() int64 {
	item.RLock()
	defer item.RUnlock()
	return item.accessCount
}


func (item *CacheItem) Key() interface{} {
	// immutable
	return item.key
}

func (item *CacheItem) Data() interface{} {
	// immutable
	return item.data
}

// 在item被移除cache时 被触发的操作：由用户自定义操作
func (item *CacheItem) SetAboutToExpireCallback(f func(interface{})) {
	item.Lock()
	defer item.Unlock()
	item.aboutToExpire = f
}
