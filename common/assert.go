package common

import "fmt"

// 断言 assert
func Assert(v bool, a ...interface{}){
	if !v{
		if msg := fmt.Sprint(a...); msg != "" {
			panic(fmt.Sprintf("assert failed, %s!", msg))
		} else {
			panic(fmt.Sprintf("assert failed!"))
		}
	}
}
