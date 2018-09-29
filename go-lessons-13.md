# Go 并发资源竞争
有并发，就有资源竞争，如果两个或者多个goroutine在没有相互同步的情况下，访问某个共享的资源，比如同时对该资源进行读写时，就会处于相互竞争的状态，这就是并发中的资源竞争。

并发本身并不复杂，但是因为有了资源竞争的问题，就使得我们开发出好的并发程序变得复杂起来，因为会引起很多莫名其妙的问题。
~~~
package main
import (
    "fmt"
    "runtime"
    "sync"
)
var (
    count int32
    wg    sync.WaitGroup
)
func main() {
    wg.Add(2)
    go incCount()
    go incCount()
    wg.Wait()
    fmt.Println(count)
}
func incCount() {
    defer wg.Done()
    for i := 0; i < 2; i++ {
        value := count
        runtime.Gosched()
        value++
        count = value
    }
}
~~~
这是一个资源竞争的例子，我们可以多运行几次这个程序，会发现结果可能是2，也可以是3，也可能是4。因为共享资源count变量没有任何同步保护，所以两个goroutine都会对其进行读写，会导致对已经计算好的结果覆盖，以至于产生错误结果，这里我们演示一种可能，两个goroutine我们暂时称之为g1和g2。
> 1、g1读取到count为0。
   2、然后g1暂停了，切换到g2运行，g2读取到count也为0。
   3、g2暂停，切换到g1，g1对count+1，count变为1。
  4、g1暂停，切换到g2，g2刚刚已经获取到值0，对其+1，最后赋值给count还是1

有没有注意到，刚刚g1对count+1的结果被g2给覆盖了，两个goroutine都+1还是1
不再继续演示下去了，到这里结果已经错了，两个goroutine相互覆盖结果。我们这里的runtime.Gosched()是让当前goroutine暂停的意思，退回执行队列，让其他等待的goroutine运行，目的是让我们演示资源竞争的结果更明显。注意，这里还会牵涉到CPU问题，多核会并行，那么资源竞争的效果更明显。

所以我们对于同一个资源的读写必须是原子化的，也就是说，同一时间只能有一个goroutine对共享资源进行读写操作。

### 共享资源
共享资源竞争的问题，非常复杂，并且难以察觉，好在Go为我们提供了一个工具帮助我们检查，这个就是go build -race命令。我们在当前项目目录下执行这个命令，生成一个可以执行文件，然后再运行这个可执行文件，就可以看到打印出的检测信息。
~~~
go build -race
~~~
多加了一个-race标志，这样生成的可执行程序就自带了检测资源竞争的功能，下面我们运行，也是在终端运行。
~~~
./hello
~~~
我这里示例生成的可执行文件名是hello，所以是这么运行的，这时候，我们看终端输出的检测结果。
~~~
➜  hello ./hello       
==================
WARNING: DATA RACE
Read at 0x0000011a5118 by goroutine 7:
  main.incCount()
      /Users/xxx/code/go/src/flysnow.org/hello/main.go:25 +0x76

Previous write at 0x0000011a5118 by goroutine 6:
  main.incCount()
      /Users/xxx/code/go/src/flysnow.org/hello/main.go:28 +0x9a

Goroutine 7 (running) created at:
  main.main()
      /Users/xxx/code/go/src/flysnow.org/hello/main.go:17 +0x77

Goroutine 6 (finished) created at:
  main.main()
      /Users/xxx/code/go/src/flysnow.org/hello/main.go:16 +0x5f
==================
4
Found 1 data race(s)
~~~
看，找到一个资源竞争，连在那一行代码出了问题，都标示出来了。goroutine 7在代码25行读取共享资源value := count,而这时goroutine 6正在代码28行修改共享资源count = value,而这两个goroutine都是从main函数启动的，在16、17行，通过go关键字。

既然我们已经知道共享资源竞争的问题，是因为同时有两个或者多个goroutine对其进行了读写，那么我们只要保证，同时只有一个goroutine读写不就可以了，现在我们就看下传统解决资源竞争的办法—对资源加锁。

Go语言提供了atomic包和sync包里的一些函数对共享资源同步枷锁，我们先看下atomic包。
~~~
package main
import (
    "fmt"
    "runtime"
    "sync"
    "sync/atomic"
)
var (
    count int32
    wg    sync.WaitGroup
)
func main() {
    wg.Add(2)
    go incCount()    
    go incCount()
    wg.Wait()
    fmt.Println(count)
}
func incCount() {
    defer wg.Done()    
    for i := 0; i < 2; i++ {
        value := atomic.LoadInt32(&count)
        runtime.Gosched()
        value++
        atomic.StoreInt32(&count,value)
    }
}
~~~
留意这里atomic.LoadInt32和atomic.StoreInt32两个函数，一个读取int32类型变量的值，一个是修改int32类型变量的值，这两个都是原子性的操作，Go已经帮助我们在底层使用加锁机制，保证了共享资源的同步和安全，所以我们可以得到正确的结果，这时候我们再使用资源竞争检测工具go build -race检查，也不会提示有问题了。

atomic包里还有很多原子化的函数可以保证并发下资源同步访问修改的问题，比如函数atomic.AddInt32可以直接对一个int32类型的变量进行修改，在原值的基础上再增加多少的功能，也是原子性的，这里不再举例，大家自己可以试试。

atomic虽然可以解决资源竞争问题，但是比较都是比较简单的，支持的数据类型也有限，所以Go语言还提供了一个sync包，这个sync包里提供了一种互斥型的锁，可以让我们自己灵活的控制哪些代码，同时只能有一个goroutine访问，被sync互斥锁控制的这段代码范围，被称之为临界区，临界区的代码，同一时间，只能又一个goroutine访问。刚刚那个例子，我们还可以这么改造。
~~~
package main
import (
    "fmt"
    "runtime"
    "sync"
)
var (
    count int32
    wg    sync.WaitGroup
    mutex sync.Mutex
)
func main() {
    wg.Add(2)
    go incCount()    
    go incCount()
    wg.Wait()
    fmt.Println(count)
}
func incCount() {
    defer wg.Done()    
    for i := 0; i < 2; i++ {
        mutex.Lock()
        value := count
        runtime.Gosched()
        value++
        count = value
        mutex.Unlock()
    }
}
~~~
实例中，新声明了一个互斥锁mutex sync.Mutex，这个互斥锁有两个方法，一个是mutex.Lock(),一个是mutex.Unlock(),这两个之间的区域就是临界区，临界区的代码是安全的。

示例中我们先调用mutex.Lock()对有竞争资源的代码加锁，这样当一个goroutine进入这个区域的时候，其他goroutine就进不来了，只能等待，一直到调用mutex.Unlock() 释放这个锁为止。

这种方式比较灵活，可以让代码编写者任意定义需要保护的代码范围，也就是临界区。除了原子函数和互斥锁，Go还为我们提供了更容易在多个goroutine同步的功能，这就是通道chan
