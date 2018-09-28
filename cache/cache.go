package cache

import (
	"io"
	)

// Cache is thread-safe cache
// See https://github.com/google/leveldb/blob/master/include/leveldb/cache.h
// 定义Cache接口：insert、lookup、erase等行为
type Cache interface {
	// 返回一个新的数值ID。
	// 常用于多clients共享相同cache来分片key空间
	// 当client启动时分配一个新的ID,并可以采用id_key的形式
	NewId()	uint64

	//  插入: 建立一个从key-value到cache的映射，同时分配指定size相当于cache的总容量(capacity)
	//  返回handle(相当于mapping【key:value ---> cache】).
	// 注：当返回的handle mapping不再需要，调用方必须调用handle.Close()
	//     当插入entry不再需要的时候，key-value会传递给“deleter”，由调用方处理
	Insert(key string, value interface{}, size int, deleter func(key string, value interface{})) (handle io.Closer)


	// 查询
	// 若是cache没有key的mapping，则直接返回 nil, nil, false
	// 否则返回 value，handle，true
	// 注：当返回的handle mapping不再需要，调用方必须调用handle.Close()
	Lookup(key string) (value interface{}, handle io.Closer, ok bool)


	// 删除
	// 根据指定key删除指定内容value及mapping
	// 注：key所关联的entry必须当所有已存在的handle释放了方可被删除
    Erase(key string)

	// 清理cache
	// 通过调用“deleter”函数来清理所有exist的entry
	// 注：必须所有handle被释放 方可完成释放
	Close()
}

// 根据指定capacity创建cache
func New(capacity int64) Cache{
	return NewLRUCache(capacity)
}
