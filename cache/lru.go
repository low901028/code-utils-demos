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

	runtime.SetFinalizer(p, (*_LRUCache).Close)  // 退出清理Cache
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

// 设置
func (p *LRUCache) Set(key string, value interface{}, size int, deleter ...func(key string, value interface{})) {
	if len(deleter) > 0 {
		h := p.Insert(key, value, size, deleter[0])
		h.Close()
	} else {
		h := p.Insert(key, value, size, nil)
		h.Close()
	}
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
	if element == nil{
		return nil, nil, false
	}

	// 若是存在 则将element放置到表头
	p.list.MoveToFront(element)
	h := element.Value.(*LRUHandle)
	h.time_accessed.Store(time.Now())
	p.addref(h)

	return h.Value(), h, true
}

// 获取cache中key对应的内容 并删除双向链表和hash table中的记录
func (p *LRUCache) Take(key string) (handle io.Closer, ok bool){
	p.mu.Lock()
	defer p.mu.Unlock()

	element := p.table[key]
	if element == nil{
		return nil, false
	}

	p.list.Remove(element)
	delete(p.table, key)

	h := element.Value.(*LRUHandle)

	return h, true
}


// 功能很类似Take 额外需要release对应的key关联的handle
func (p *LRUCache) Erase(key string){
	p.mu.Lock()
	defer p.mu.Unlock()

	element := p.table[key]
	if element != nil{
		return
	}

	p.list.Remove(element)
	delete(p.table, key)

	h := element.Value.(*LRUHandle)
	p.unref(h)   // 删除key  需要release关联的handle

	return
}


// 设置cache的capacity
func (p *LRUCache) SetCapacity(capacity int64){
	p.mu.Lock()
	defer p.mu.Unlock()

	common.Assert(capacity > 0)
	p.capacity = capacity
	p.checkCapacity()  // 检查size 是否超过capacity

}


// 统计信息cache
func (p *LRUCache) Stats() (length, size, capacity int64, oldest time.Time){
	p.mu.Lock()
	defer p.mu.Unlock()

	if lastElem := p.list.Back(); lastElem != nil{
		oldest = lastElem.Value.(*LRUHandle).time_accessed.Load().(time.Time)
	}
	return int64(p.list.Len()), p.size, p.capacity, oldest
}
// 统计信息json格式
func (p *LRUCache) StatsJSON() string {
	if p == nil {
		return "{}"
	}
	l, s, c, o := p.Stats()
	return fmt.Sprintf(`{
	"Length": %v,
	"Size": %v,
	"Capacity": %v,
	"OldestAccess": "%v"
}`, l, s, c, o)
}

// cache中element的个数
func (p *LRUCache) Length() int64{
	p.mu.Lock()
	defer p.mu.Unlock()

	return int64(p.list.Len())
}

//
func (p *LRUCache) Size() int64{
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.size
}

func (p *LRUCache) Capacity() int64{
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.capacity
}

// cache中最新element对应的时间
// 若是cache不存在 则返回IsZero() time
func (p *LRUCache) Newest() (newest time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if frontElem := p.list.Front(); frontElem != nil {
		newest = frontElem.Value.(*LRUHandle).time_accessed.Load().(time.Time)
	}
	return
}

func (p *LRUCache) Oldest() (oldest time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if lastElem := p.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*LRUHandle).time_accessed.Load().(time.Time)
	}
	return
}


// cache中所有的keys； 按照使用时间的最近进行排序
func (p *LRUCache) Keys() []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	keys := make([]string, 0, p.list.Len())
	for e := p.list.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*LRUHandle).key)
	}
	return keys
}

// 清除cache
// 前提要release所有key关联的handle
func (p *LRUCache) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, element := range p.table {
		h := element.Value.(*LRUHandle)
		p.unref(h)
	}

	p.list = list.New()
	p.table = make(map[string]*list.Element)
	p.size = 0
	return
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

//==========================================实现io.Closer==========================================
func (p *_LRUCache) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, element := range p.table {
		h := element.Value.(*LRUHandle)
		common.Assert(h.refs == 1, "h.refs = ", h.refs)
		p.unref(h)
	}

	p.list = nil
	p.table = nil
	p.size = 0
}


//==========================================cache lur实现扩展==========================================

// 查询二级索引hash表判断对应的key是否存在
func (p *LRUCache) HashKey(key string) bool{
	p.mu.Lock()
	defer p.mu.Unlock()

	_, ok := p.table[key]
	return ok
}

// 先获取双向链表表头的element 拿到关联的handle
// 接着从handle拿到key
func (p *LRUCache) FrontKey() (key string){
	if h := p.Front(); h != nil{
		key = h.Key()
		h.Close()

		return key
	}
	return ""
}

//同FrontKey
func (p *LRUCache) BackKey() (key string) {
	if h := p.Back(); h != nil {
		key = h.Key()
		h.Close()
		return key
	}
	return ""
}

// 同FrontKey 先拿到双向链表表头element的handle
// 不过当handle没有内容时，则可使用defaultvalues[0]作为结果
func (p *LRUCache) FrontValue(defaultValue ...interface{}) (value interface{}) {
	if h := p.Front(); h != nil {
		value = h.Value()
		h.Close()
		return
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	} else {
		return nil
	}
}

// 同FrontValue
func (p *LRUCache) BackValue(defaultValue ...interface{}) (value interface{}) {
	if h := p.Back(); h != nil {
		value = h.Value()
		h.Close()
		return
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	} else {
		return nil
	}
}

// 移除双向链表表头
func (p *LRUCache) RemoveFront() {
	if h := p.PopFront(); h != nil {
		h.Close()
	}
}

// 移除双向链表表尾
func (p *LRUCache) RemoveBack() {
	if h := p.PopBack(); h != nil {
		h.Close()
	}
}

// 先获取双向列表表头的element  获取到关联的handle【LRUHandle类似redis里面的redisobject】
func (p *LRUCache) Front() (h *LRUHandle) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var element *list.Element
	if element = p.list.Front(); element == nil {
		return
	}

	h = element.Value.(*LRUHandle)
	p.addref(h)   // 使用handle 一定要增加ref数
	return
}

// 同Front
func (p *LRUCache) Back() (h *LRUHandle) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var element *list.Element
	if element = p.list.Back(); element == nil {
		return
	}

	h = element.Value.(*LRUHandle)
	p.addref(h)
	return
}

// 将element压入到表头
func (p *LRUCache) PushFront(key string, value interface{}, size int, deleter func(key string, value interface{})) {
	p.mu.Lock()
	defer p.mu.Unlock()

	common.Assert(key != "" && size > 0)
	if element := p.table[key]; element != nil {   // 添加element已存在，则需要指定清理操作：双向链表remove  二级索引table delete
		p.list.Remove(element)
		delete(p.table, key)

		h := element.Value.(*LRUHandle)
		p.unref(h)
	}

	h := &LRUHandle{
		c:            p,
		key:          key,
		value:        value,
		size:         int64(size),
		deleter:      deleter,
		time_created: time.Now(),
		refs:         1, // 添加element至少会产生一个ref【此处没有返回handle 故而只有一个ref】
	}
	h.time_accessed.Store(time.Now())

	element := p.list.PushFront(h)
	p.table[key] = element
	p.size += h.size
	p.checkCapacity()
	return
}

// 同PushFront：将element压入到双向链表的表尾
func (p *LRUCache) PushBack(key string, value interface{}, size int, deleter func(key string, value interface{})) {
	p.mu.Lock()
	defer p.mu.Unlock()

	common.Assert(key != "" && size > 0)
	if element := p.table[key]; element != nil {
		p.list.Remove(element)
		delete(p.table, key)

		h := element.Value.(*LRUHandle)
		p.unref(h)
	}

	h := &LRUHandle{
		c:            p,
		key:          key,
		value:        value,
		size:         int64(size),
		deleter:      deleter,
		time_created: time.Now(),
		refs:         1, //添加element至少会产生一个ref【此处没有返回handle 故而只有一个ref】
	}
	h.time_accessed.Store(time.Now())

	element := p.list.PushBack(h)
	p.table[key] = element
	p.size += h.size
	p.checkCapacity()
	return
}

// 弹出双向链表尾element
func (p *LRUCache) PopBack() (h *LRUHandle) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var element *list.Element
	if element = p.list.Back(); element == nil {
		return
	}

	h = element.Value.(*LRUHandle)
	delete(p.table, h.Key())
	p.list.Remove(element)
	return
}

// 弹出双向链表=头element
func (p *LRUCache) PopFront() (h *LRUHandle) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var element *list.Element
	if element = p.list.Front(); element == nil {
		return
	}

	h = element.Value.(*LRUHandle)
	delete(p.table, h.Key())
	p.list.Remove(element)
	return
}

// 移动到双向链表表头
func (p *LRUCache) MoveToFront(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	element := p.table[key]
	if element == nil {
		return
	}

	p.list.MoveToFront(element)
	return
}

// 移动到双向链表表尾
func (p *LRUCache) MoveToBack(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	element := p.table[key]
	if element == nil {
		return
	}

	p.list.MoveToBack(element)
	return
}











