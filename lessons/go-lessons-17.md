# Go 读写锁 
前面的有篇文章在讲资源竞争的时候，讲互斥锁，互斥锁的根本就是当一个goroutine访问的时候，其他goroutine都不能访问，这样肯定保证了资源的同步，避免了竞争，不过也降低了性能。

仔细剖析我们的场景，当我们读取一个数据的时候，如果这个数据永远不会被修改，那么其实是不存在资源竞争的问题的，因为数据是不变的，不管怎么读取，多少goroutine同时读取，都是可以的。

所以其实读取并不是问题，问题主要是修改，修改的数据要同步，这样其他goroutine才可以感知到。所以真正的互斥应该是读取和修改、修改和修改之间，读取和读取是没有互斥操作的。

所以这就延伸出来另外一种锁，叫做读写锁。

读写锁可以让多个读操作同时并发，同时读取，但是对于写操作是完全互斥的。也就是说，当一个goroutine进行写操作的时候，其他goroutine既不能进行读操作，也不能进行写操作。
~~~
var count int
var wg sync.WaitGroup
func main() {
    wg.Add(10)
    for i:=0;i<5;i++ {
            go read(i)
    }    
    for i:=0;i<5;i++ {
            go write(i);
    }

    wg.Wait()
}
func read(n int) {
    fmt.Printf("读goroutine %d 正在读取...\n",n)

    v := count

    fmt.Printf("读goroutine %d 读取结束，值为：%d\n", n,v)
    wg.Done()
}
func write(n int) {
    fmt.Printf("写goroutine %d 正在写入...\n",n)
    v := rand.Intn(1000)

    count = v

    fmt.Printf("写goroutine %d 写入结束，新值为：%d\n", n,v)
    wg.Done()
}
~~~
以上我们定义了一个共享的资源count，并且声明了2个函数进行读写read和write，在main函数的测试中，我们同时启动了5个读写goroutine进行读写操作，通过打印的结果来看，写入操作是处于竞争状态的，有的写入操作被覆盖了。通过go build -race也可以看到更明细的竞争态。

针对这种情况，第一个方案是加互斥锁，同时智能有一个goroutine可以操作count，但是这种方法性能比较慢，而且我们说的读操作可以不互斥，所以这种情况比较适合使用读写锁。
~~~
var count int
var wg sync.WaitGroup
var rw sync.RWMutex
func main() {
    wg.Add(10)
    for i:=0;i<5;i++ {
            go read(i)
    }    
    for i:=0;i<5;i++ {
            go write(i);
    }

    wg.Wait()
}
func read(n int) {
    rw.RLock()
    fmt.Printf("读goroutine %d 正在读取...\n",n)

    v := count

    fmt.Printf("读goroutine %d 读取结束，值为：%d\n", n,v)
    wg.Done()
    rw.RUnlock()
}
func write(n int) {
    rw.Lock()
    fmt.Printf("写goroutine %d 正在写入...\n",n)
    v := rand.Intn(1000)

    count = v

    fmt.Printf("写goroutine %d 写入结束，新值为：%d\n", n,v)
    wg.Done()
    rw.Unlock()
}
~~~
我们在read里使用读锁，也就是RLock和RUnlock，写锁的方法名和我们平时使用的一样Lock和Unlock,这样，我们就使用了读写锁，可以并发的读，但是同时只能有一个写，并且写的时候不能进行读操作，现在我们再运行代码，可以从输出的数据看到，可以独到新值了。

我们同时也可以使用go build -race 检测，也没有竞争提示了。

我们在做Java开发的时候，肯定知道SynchronizedMap这个Map，它是一个在多线程下安全的Map，我们可以通过Collections.synchronizedMap(Map<K, V>)来获取一个安全的Map，下面我们看看如何使用读写锁，基于Go语言来实现一个安全的Map 。
~~~
package common
import (
    "sync"
)
//安全的Map
type SynchronizedMap struct {
    rw *sync.RWMutex
    data map[interface{}]interface{}
}
//存储操作
func (sm *SynchronizedMap) Put(k,v interface{}){
    sm.rw.Lock()
    defer sm.rw.Unlock()

    sm.data[k]=v
}
//获取操作
func (sm *SynchronizedMap) Get(k interface{}) interface{}{
    sm.rw.RLock()
    defer sm.rw.RUnlock()
    
    return sm.data[k]
}
//删除操作
func (sm *SynchronizedMap) Delete(k interface{}) {
    sm.rw.Lock()
    defer sm.rw.Unlock()
    
    delete(sm.data,k)
}
//遍历Map，并且把遍历的值给回调函数，可以让调用者控制做任何事情
func (sm *SynchronizedMap) Each(cb func (interface{},interface{})){
    sm.rw.RLock()
    defer sm.rw.RUnlock()
    for k, v := range sm.data {
        cb(k,v)
    }
}
//生成初始化一个SynchronizedMap
func NewSynchronizedMap() *SynchronizedMap{
    return &SynchronizedMap{
        rw:new(sync.RWMutex),
        data:make(map[interface{}]interface{}),
    }
}
~~~
这个安全的Map被我们定义为一个SynchronizedMap的结构体，这个结构体里有两个字段，一个是读写锁rw,一个是存储数据的data，data是map类型。

然后就是给SynchronizedMap定义一些方法，如果这些方法是增删改的，就要使用写锁，如果是只读的，就使用读锁，这样就保证了我们数据data在多个goroutine下的安全性。

有了这个安全的Map我们就可以在多goroutine下增删改查数据了，都是安全的。

这里定义了一个Each方法，这个方法很有意思，用过Gradle的都知道，也有类似遍历Map的方法。这个方法我们可以传入一个回调函数作为参数，来对我们遍历的SynchronizedMap数据进行处理，比如我打印SynchronizedMap中的数据。
~~~
sm.Each(func(k interface{}, v interface{}) {
    fmt.Println(k," is ",v)
})
~~~
sm就是一个SynchronizedMap，非常简洁吧。

以上就是读写锁使用的一个例子，我们可以把这个map数据当成缓存数据，或者当成数据库，然后使用读写锁进行控制，可以多读，但是只能有一个写。
