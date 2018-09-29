package main

import (
	"fmt"
	"code-utils-demos/cache"
	"log"
)

func main() {
	c := cache.NewLRUCache(100)
	defer c.Close()
	//===========================简单操作:Set、 Value=========================
	c.Set("key1", "value1", 1)
	value1 := c.Value("key1").(string)
	fmt.Println("key1:", value1)

	c.Set("key1","value11111",1)
	value1 = c.Value("key1").(string)
	fmt.Println("key1: ", value1)

	c.Set("key2", "value2", 1)
	value2 := c.Value("key2", "null").(string)
	fmt.Println("key2:", value2)

	value3 := c.Value("key3", "null").(string)
	fmt.Println("key3:", value3)

	value4 := c.Value("key4") // value4 is nil
	fmt.Println("key4:", value4)

	fmt.Println("==============================Done==========================")

	//===========================简单操作：newId、Insert、Lookup、Erase、=========================
	cc := cache.NewLRUCache(10)
	defer cc.Close()
	// 创建new id
	id0 := cc.NewId()
	id1 := cc.NewId()
	id2 := cc.NewId()
	fmt.Println("id0:", id0)
	fmt.Println("id1:", id1)
	fmt.Println("id2:", id2)

	// insert
	v1 := "data:123"
	h1 := cc.Insert("123", "data:123", len("data:123"), func(key string, value interface{}) {
		fmt.Printf("deleter(%q:%q)\n", key, value)
	})

	// lookup
	v2, h2, ok := cc.Lookup("123")
	assert(ok)
	assert(h2 != nil)

	// remove
	cc.Erase("123")

	// lookup：key不存在
	_, h3, ok := cc.Lookup("123")
	assert(!ok)
	assert(h3 == nil)

	//
	fmt.Printf("user1(%s)\n", v1)
	fmt.Printf("user2(%s)\n", v2.(string))

	// release h1
	// because the h2 handle the value, so the deleter is not ivoked!
	// 此处虽然relaease h1 不过由于h2关联着底层handle，会导致deleter在此处不起作用的
	// TODO：试试同时释放h1/h2 看看
	h1.Close()

	// 以下代码会触发deleter操作  由于对应的key=(123)  handle没有外部引用了(refs = 0) 则会触发deleter操作
	fmt.Println("invoke deleter(123) begin")
	h2.Close()
	fmt.Println("invoke deleter(123) end")

	// insert
	h4 := cc.Insert("abc", "data:abc", len("data:abc"), func(key string, value interface{}) {
		fmt.Printf("deleter(%q:%q)\n", key, value)
	})
	// release h4
	// 此处虽然释放了handle 但是cache仍持有该key的handle，默认情况下新建的key对应的handle refs = 2 具体原因见lru.go
	h4.Close()

	// cache length
	length := cc.Length()
	assert(length == 1)

	// cache size
	size := cc.Size()
	assert(size == 8, "size:", size)

	// add h5
	// this will cause the capacity(10) overflow, so the h4 deleter will be invoked
	// 以下代码会触发h4 deleter: 由于当前的cache size > capacity 导致会触发压缩行为  淘汰时间相对久的key 故而会引发h4 deleter
	fmt.Println("invoke deleter(h4) begin")
	h5 := cc.Insert("456", "data:456", len("data:456"), func(key string, value interface{}) {
		fmt.Printf("deleter(%q:%q)\n", key, value)
	})
	fmt.Println("invoke deleter(h4) end")

	// 释放所有的handle
	h5.Close()

	// 统计
	fmt.Println("StatsJSON:", cc.StatsJSON())

	// done
	fmt.Println("========================== Done 2.0 ===========================")

	// ========================================= 简单操作：LRUHandle========================================
	ccc := cache.NewLRUCache(100)
	defer ccc.Close()

	h11 := ccc.Insert("100", "101", 1, func(key string, value interface{}) {
		fmt.Printf("deleter(%q, %q)\n", key, value.(string))
	})
	v11 := h11.(*cache.LRUHandle).Value().(string)
	fmt.Printf("v1: %s\n", v11)
	h11.Close()

	_, h22, ok := ccc.Lookup("100")
	if !ok {
		log.Fatal("lookup failed!")
	}
	defer h22.Close()

	// h2 still valid after Erase
	ccc.Erase("100")
	v22 := h22.(*cache.LRUHandle).Value().(string)
	fmt.Printf("v2: %s\n", v22)

	// but new lookup will failed
	if _, _, ok := ccc.Lookup("100"); ok {
		log.Fatal("lookup succeed!")
	}

	fmt.Println("================================= Done ===========================")
}

//
func assert(v bool, a ...interface{}) {
	if !v {
		panic(fmt.Sprint(a...))
	}
}