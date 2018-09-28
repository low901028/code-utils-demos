package main

import (
	"fmt"
	"code-utils-demos/cache"
)

func main() {
	c := cache.NewLRUCache(100)
	defer c.Close()
	//===========================简单操作=========================
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

	fmt.Println("Done")

	//===========================简单操作=========================
}
