package cache

import (
	"sync"
	"container/list"
	"time"
	"sync/atomic"
	"code-utils-demos/common"
	"runtime"
	"fmt"
	"io"
)

var _Cache = (*LRUCache)(nil)

type LRUCache struct {
	*_LRUCache
}

type _LRUCache struct {
	// 锁： 控制cache线程安全的保障
	mu sync.Mutex

	// 双向链表
	list 	*list.List
	// hash表映射key-value
	table 	map[string]*list.Element

	// 当前cache的size
	size int64

	// cache的容量【固定】
	capacity  int64

	//
	last_id uint64
}

// 包装key-value存在cache【LRUCache】
type LRUHandle struct {
	c				*LRUCache
	key 			string
	value			interface{}
	size  			int64
	deleter			func(key string, value interface{})
	time_created	time.Time
	time_accessed	atomic.Value
	refs			uint32
}

// ========================================LRUHandle=====================================

//
func (h *LRUHandle) Key()	string{
	return h.key
}

//
func (h *LRUHandle)	Value() interface{}{
	return h.value
}


func (h *LRUHandle)	Size() int{
	return int(h.size)
}


func (h *LRUHandle) TimeCreated() time.Time{
	return h.time_created
}


func (h *LRUHandle) Time_Accessed()	time.Time{
	return h.time_accessed.Load().(time.Time)
}


func (h *LRUHandle) Retain() (handle *LRUHandle){
	h.c.mu.Lock()
	defer h.c.mu.Unlock()
	h.c.addref(h)
	return h
}

//
func (h *LRUHandle) Close() error{
	h.c.mu.Lock()
	defer h.c.mu.Unlock()
	h.c.unref(h)
	return nil
}

// ========================================LRUCache=====================================
// 创建LRU cache
func NewLRUCache(capacity int64) *LRUCache{
	common.Assert(capacity > 0)

	p := &_LRUCache{
		list: list.New(),
		table: make(map[string]*list.Element),
		capacity:	capacity,
	}

	runtime.SetFinalizer(p, (*_LRUCache).Close)
	return &LRUCache{p}
}

// 关闭cache
func (p *LRUCache) Close() error{
	runtime.SetFinalizer(p._LRUCache, nil)
	p._LRUCache.Close()
	return nil
}

// 查询
func (p *LRUCache) Get(key string) (value interface{}, ok bool){
	if v, h, ok := p.Lookup(key); ok{
		h.Close()
		return  v, ok
	}
	return
}

// 若cache中存在 则直接获取
// 否则通过getter获取 并将获取的内容set到cache
func (p *LRUCache) GetFrom(key string, getter func(key string) (v interface{} , size int , err error)) (value interface{} , err error){
	if v, h, ok := p.Lookup(key); ok{  // cache中存在
		h.Close()
		return v, nil
	}

	if getter == nil{
		return nil, fmt.Errorf("cache: %q not found!", key)
	}

	value, size, err := getter(key)
	if err != nil{
		return
	}

	common.Assert(size > 0)
	p.Set(key, value, size)

	return
}

// 获取value
// 若是cache中没有对应的value，取defaultValue第一个元素作为结果返回
func (p *LRUCache) Value(key string, defaultValue ...interface{}) interface{}{
	if v, h, ok := p.Lookup(key); ok{
		h.Close()
		return v
	}

	if len(defaultValue) > 0{
		return defaultValue[0]
	} else{
		return nil
	}
}

//
func (p *LRUCache) NewId() uint64{
	p.mu.Lock()
	defer p.mu.Unlock()

	p.last_id++
	return p.last_id
}


// 插入
func (p *LRUCache) Insert(key string, value interface{}, size int, deleter func(key string , value interface{})) (handle io.Closer){
	handle = p.Insert_(key, value, size, deleter)
	return
}

func (p *LRUCache) Insert_(key string, value interface{}, size int, deleter func(key string, value interface{})) (handle *LRUHandle){
	p.mu.Lock()
	defer p.mu.Unlock()

	common.Assert(key != "" && size > 0)

	if element := p.table[key]; element != nil{
		p.list.Remove(element)
		delete(p.table, key)

		h := element.Value.(*LRUHandle)
		p.unref(h)
	}

	h := &LRUHandle{
		c:				p,
		key:			key,
		value:			value,
		size:			int64(size),
		deleter:		deleter,
		time_created: 	time.Now(),
		refs:			2,  // 1 ---> LRUCache   2 ----> 返回值handle
	}
	h.time_accessed.Store(time.Now())

	element := p.list.PushFront(h)   // 最新的数据都在表头
	p.table[key] = element
	p.size += h.size
	p.checkCapacity()                // 添加cache时  需要检查cache的capacity是否已满(size > capacity) 若已满需进行压缩
	return  h
}


// 查询
func (p *LRUCache) Lookup(key string) (value interface{}, handle io.Closer, ok bool){
	if v, h, ok := p.Lookup_(key); ok{
		return v, h, ok
	}
	return
}

func (p *LRUCache) Lookup_(key string) (value interface{}, handle *LRUHandle, ok bool){
	p.mu.Lock()
	defer p.mu.Unlock()

	element := p.table[key]  // 先从二级索引hash table拿数据  若是没有也意味双向链表也没有
	if element != nil{
		return nil, nil, false
	}

	// 若是存在 则将element放置到表头
	p.list.MoveToFront(element)
	h := element.Value.(*LRUHandle)
	h.time_accessed.Store(time.Now())
	p.addref(h)

	return h.Value(), h, true
}

// 检查cache的size是否已经超过capacity
// 一旦超过了 则进行收缩： 淘汰旧数据 直至size <= capacity
func (p *LRUCache) checkCapacity() {
	for p.size > p.capacity && len(p.table) > 1 {
		delElem := p.list.Back()
		h := delElem.Value.(*LRUHandle)
		p.list.Remove(delElem)
		delete(p.table, h.key)
		p.unref(h)
	}
}

func (p *_LRUCache) addref(h *LRUHandle) {
	h.refs++
}

func (p *_LRUCache) unref(h *LRUHandle) {
	common.Assert(h.refs > 0)
	h.refs--
	if h.refs <= 0 {
		p.size -= h.size
		if h.deleter != nil {
			h.deleter(h.key, h.value)
		}
	}
}















