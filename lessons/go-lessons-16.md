# Go并发示例-Pool
这篇文章演示使用有缓冲的通道实现一个资源池，这个资源池可以管理在任意多个goroutine之间共享的资源，比如网络连接、数据库连接等，我们在数据库操作的时候，比较常见的就是数据连接池，也可以基于我们实现的资源池来实现。

可以看出，资源池也是一种非常流畅性的模式，这种模式一般适用于在多个goroutine之间共享资源，每个goroutine可以从资源池里申请资源，使用完之后再放回资源池里，以便其他goroutine复用。

好了，老规矩，我们先构建一个资源池结构体，然后再赋予一些方法，这个资源池就可以帮助我们管理资源了。
~~~
//一个安全的资源池，被管理的资源必须都实现io.Close接口

type Pool struct {
    m sync.Mutex
    res chan io.Closer
    factory func() (io.Closer,error)
    closed bool
}
~~~
这个结构体Pool有四个字段，其中m是一个互斥锁，这主要是用来保证在多个goroutine访问资源时，池内的值是安全的。

- res字段是一个有缓冲的通道，用来保存共享的资源，这个通道的大小，在初始化Pool的时候就指定的。注意这个通道的类型是io.Closer接口，所以实现了这个io.Closer接口的类型都可以作为资源，交给我们的资源池管理。

- factory这个是一个函数类型，它的作用就是当需要一个新的资源时，可以通过这个函数创建，也就是说它是生成新资源的，至于如何生成、生成什么资源，是由使用者决定的，所以这也是这个资源池灵活的设计的地方。

- closed字段表示资源池是否被关闭，如果被关闭的话，再访问是会有错误的。

现在先这个资源池我们已经定义好了，也知道了每个字段的含义，下面就开时具体使用。刚刚我们说到关闭错误，那么我们就先定义一个资源池已经关闭的错误。
~~~
var ErrPoolClosed = errors.New("资源池已经关闭。")
~~~
非常简洁，当我们从资源池获取资源的时候，如果该资源池已经关闭，那么就会返回这个错误。单独定义它的目的，是和其他错误有一个区分，这样需要的时候，我们就可以从众多的error类型里区分出来这个ErrPoolClosed。

下面我们就该为创建Pool专门定一个函数了，这个函数就是工厂函数，我们命名为New。
~~~
//创建一个资源池
func New(fn func() (io.Closer, error), size uint) (*Pool, error) {
    if size <= 0 {
            return nil, errors.New("size的值太小了。")
    }   
    return &Pool{
        factory: fn,
        res:     make(chan io.Closer, size),
    }, nil
}
~~~
这个函数创建一个资源池，它接收两个参数，一个fn是创建新资源的函数；还有一个size是指定资源池的大小。

这个函数里，做了size大小的判断，起码它不能小于或者等于0，否则就会返回错误。如果参数正常，就会使用size创建一个有缓冲的通道，来保存资源，并且返回一个资源池的指针。

有了创建好的资源池，那么我们就可以从中获取资源了。
~~~
//从资源池里获取一个资源
func (p *Pool) Acquire() (io.Closer,error) {
    select {
        case r,ok := <-p.res:
                log.Println("Acquire:共享资源")
                if !ok {
                            return nil,ErrPoolClosed
                }        
                return r,nil
        default:
                log.Println("Acquire:新生成资源")        
                return p.factory()
    }
}
~~~
Acquire方法可以从资源池获取资源，如果没有资源，则调用factory方法生成一个并返回。

这里同样使用了select的多路复用，因为这个函数不能阻塞，可以获取到就获取，不能就生成一个。

这里的新知识是通道接收的多参返回，如果可以接收的话，第一参数是接收的值，第二个表示通道是否关闭。例子中如果ok值为false表示通道关闭，如果为true则表示通道正常。所以我们这里做了一个判断，如果通道关闭的话，返回通道关闭错误。

有获取资源的方法，必然还有对应的释放资源的方法，因为资源用完之后，要还给资源池，以便复用。在讲解释放资源的方法前，我们先看下关闭资源池的方法，因为释放资源的方法也会用到它。

关闭资源池，意味着整个资源池不能再被使用，然后关闭存放资源的通道，同时释放通道里的资源。
~~~
//关闭资源池，释放资源
func (p *Pool) Close() {
    p.m.Lock()    
    defer p.m.Unlock()    
    if p.closed {        
          return
     }

    p.closed = true

    //关闭通道，不让写入了
    close(p.res)    //关闭通道里的资源
    for r:=range p.res {
        r.Close()
    }
}
~~~
这个方法里，我们使用了互斥锁，因为有个标记资源池是否关闭的字段closed需要再多个goroutine操作，所以我们必须保证这个字段的同步。这里把关闭标志置为true。

然后我们关闭通道，不让写入了，而且我们前面的Acquire也可以感知到通道已经关闭了。同比通道后，就开始释放通道中的资源，因为所有资源都实现了io.Closer接口，所以我们直接调用Close方法释放资源即可。

关闭方法有了，我们看看释放资源的方法如何实现。
~~~
func (p *Pool) Release(r io.Closer){   
     //保证该操作和Close方法的操作是安全的
    p.m.Lock()    
    defer p.m.Unlock() 
    //资源池都关闭了，就省这一个没有释放的资源了，释放即可
    if p.closed {
        r.Close()        
        return
    }
    select {
    case p.res <- r:
        log.Println("资源释放到池子里了")    
    default:
        log.Println("资源池满了，释放这个资源吧")
        r.Close()
    }
}
~~~
释放资源本质上就会把资源再发送到缓冲通道中，就是这么简单，不过为了更安全的实现这个方法，我们使用了互斥锁，保证closed标志的安全，而且这个互斥锁还有一个好处，就是不会往一个已经关闭的通道发送资源。

这是为什么呢？因为Close和Release这两个方法是互斥的，Close方法里对closed标志的修改，Release方法可以感知到，所以就直接return了，不会执行下面的select代码了，也就不会往一个已经关闭的通道里发送资源了。

如果资源池没有被关闭，则继续尝试往资源通道发送资源，如果可以发送，就等于资源又回到资源池里了；如果发送不了，说明资源池满了，该资源就无法重新回到资源池里，那么我们就把这个需要释放的资源关闭，抛弃了。

针对这个资源池管理的一步步都实现了，而且做了详细的讲解，下面就看下整个示例代码，方便理解。
~~~
package common
import (
    "errors"
    "io"
    "sync"
    "log"
)
//一个安全的资源池，被管理的资源必须都实现io.Close接口
type Pool struct {
    m       sync.Mutex
    res     chan io.Closer
    factory func() (io.Closer, error)
    closed  bool
}
var ErrPoolClosed = errors.New("资源池已经被关闭。")
//创建一个资源池
func New(fn func() (io.Closer, error), size uint) (*Pool, error) {
    if size <= 0 { 
           return nil, errors.New("size的值太小了。")
    }    
    return &Pool{
        factory: fn,
        res:     make(chan io.Closer, size),
    }, nil
}
    
//从资源池里获取一个资源
func (p *Pool) Acquire() (io.Closer,error) { 
   select {
       case r,ok := <-p.res:
        log.Println("Acquire:共享资源")       
        if !ok {
                    return nil,ErrPoolClosed
        }        
        return r,nil
    default:
        log.Println("Acquire:新生成资源")        
        return p.factory()
    }
}
 //关闭资源池，释放资源
func (p *Pool) Close() {
    p.m.Lock()
    defer p.m.Unlock()    
    if p.closed { 
           return
    }

    p.closed = true

    //关闭通道，不让写入了
    close(p.res)    //关闭通道里的资源
    for r:=range p.res {
        r.Close()
    }
}

func (p *Pool) Release(r io.Closer){
    //保证该操作和Close方法的操作是安全的
    p.m.Lock()    
    defer p.m.Unlock()   
    
     //资源池都关闭了，就省这一个没有释放的资源了，释放即可
    if p.closed {
        r.Close()        
        return 
    }    
    select {
    case p.res <- r:
        log.Println("资源释放到池子里了")    
    default:
        log.Println("资源池满了，释放这个资源吧")
        r.Close()
    }
}
~~~
好了，资源池管理写好了，也知道资源池是如何实现的啦，现在我们看看如何使用这个资源池，模拟一个数据库连接池吧。
~~~
package main
import (
    "flysnow.org/hello/common"
    "io"
    "log"
    "math/rand"
    "sync"
    "sync/atomic"
    "time"
)
const (    
    //模拟的最大goroutine
    maxGoroutine = 5
    //资源池的大小
    poolRes      = 2
)

func main() {    
    //等待任务完成
    var wg sync.WaitGroup
    wg.Add(maxGoroutine)

    p, err := common.New(createConnection, poolRes)    
    
    if err != nil {
        log.Println(err)        
        return
    }    
    //模拟好几个goroutine同时使用资源池查询数据
    for query := 0; query < maxGoroutine; query++ { 
       go func(q int) {
            dbQuery(q, p)
            wg.Done()
        }(query)
    }

    wg.Wait()
    log.Println("开始关闭资源池")
    p.Close()
}
//模拟数据库查询
func dbQuery(query int, pool *common.Pool) {
    conn, err := pool.Acquire()
    if err != nil {
        log.Println(err)       
         return
    }    
    defer pool.Release(conn)    
    //模拟查询
    time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
    log.Printf("第%d个查询，使用的是ID为%d的数据库连接", query, conn.(*dbConnection).ID)
}
//数据库连接
type dbConnection struct {
    ID int32//连接的标志
}

//实现io.Closer接口
func (db *dbConnection) Close() error {
    log.Println("关闭连接", db.ID)    
    return nil
}
var idCounter int32//生成数据库连接的方法，以供资源池使用
func createConnection() (io.Closer, error) {   
    //并发安全，给数据库连接生成唯一标志
    id := atomic.AddInt32(&idCounter, 1)    
    return &dbConnection{id}, nil
}
~~~
这时我们测试使用资源池的例子，首先定义了一个结构体dbConnection，它只有一个字段，用来做唯一标记。然后dbConnection实现了io.Closer接口，这样才可以使用我们的资源池。

createConnection函数对应的是资源池中的factory字段，用来创建数据库连接dbConnection的，同时为其赋予了一个为止的标志。

接着我们就同时开了5个goroutine，模拟并发的数据库查询dbQuery，查询方法里，先从资源池获取可用的数据库连接，用完后再释放。

这里我们会创建5个数据库连接，但是我们设置的资源池大小只有2，所以再释放了2个连接后，后面的3个连接会因为资源池满了而释放不了，一会我们看下输出的打印信息就可以看到。

最后这个资源连接池使用完之后，我们要关闭资源池，使用资源池的Close方法即可。
~~~
--------------output-------------
2017/04/17 22:25:08 Acquire:新生成资源
2017/04/17 22:25:08 Acquire:新生成资源
2017/04/17 22:25:08 Acquire:新生成资源
2017/04/17 22:25:08 Acquire:新生成资源
2017/04/17 22:25:08 Acquire:新生成资源
2017/04/17 22:25:08 第2个查询，使用的是ID为4的数据库连接
2017/04/17 22:25:08 资源释放到池子里了
2017/04/17 22:25:08 第4个查询，使用的是ID为1的数据库连接
2017/04/17 22:25:08 资源释放到池子里了
2017/04/17 22:25:08 第3个查询，使用的是ID为5的数据库连接
2017/04/17 22:25:08 资源池满了，释放这个资源吧
2017/04/17 22:25:08 关闭连接 5
2017/04/17 22:25:09 第1个查询，使用的是ID为3的数据库连接
2017/04/17 22:25:09 资源池满了，释放这个资源吧
2017/04/17 22:25:09 关闭连接 3
2017/04/17 22:25:09 第0个查询，使用的是ID为2的数据库连接
2017/04/17 22:25:09 资源池满了，释放这个资源吧
2017/04/17 22:25:09 关闭连接 2
2017/04/17 22:25:09 开始关闭资源池
2017/04/17 22:25:09 关闭连接 4
2017/04/17 22:25:09 关闭连接 1
~~~
到这里，我们已经完成了一个资源池的管理，并且进行了使用测试。
资源对象池的使用比较频繁，因为我们想把一些对象缓存起来，以便使用，这样就会比较高效，而且不会经常调用GC，为此Go为我们提供了原生的资源池管理，防止我们重复造轮子，这就是sync.Pool，我们看下刚刚我们的例子，如果用sync.Pool实现。
~~~
package main
import (
    "log"
    "math/rand"
    "sync"
    "sync/atomic"
    "time"
)
const (    
    //模拟的最大goroutine
    maxGoroutine = 5
)
func main() {
    //等待任务完成
    var wg sync.WaitGroup
    wg.Add(maxGoroutine)

    p:=&sync.Pool{
        New:createConnection,
    }   
     //模拟好几个goroutine同时使用资源池查询数据
    for query := 0; query < maxGoroutine; query++ {
      go func(q int) {
            dbQuery(q, p)
            wg.Done()
        }(query)
    }

    wg.Wait()
}
//模拟数据库查询
func dbQuery(query int, pool *sync.Pool) {
    conn:=pool.Get().(*dbConnection)    
    defer pool.Put(conn)    
    //模拟查询
    time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
    log.Printf("第%d个查询，使用的是ID为%d的数据库连接", query, conn.ID)
}
//数据库连接
type dbConnection struct {
    ID int32//连接的标志
}
//实现io.Closer接口
func (db *dbConnection) Close() error {
    log.Println("关闭连接", db.ID)
    return nil
}
var idCounter int32//生成数据库连接的方法，以供资源池使用
func createConnection() interface{} {   
    //并发安全，给数据库连接生成唯一标志
    id := atomic.AddInt32(&idCounter, 1)
    return &dbConnection{ID:id}
}
~~~
进行微小的改变即可，因为系统库没有提供New这类的工厂函数，所以我们使用字面量创建了一个sync.Pool，注意里面的New 字段，这是一个返回任意对象的方法，类似我们自己实现的资源池中的factory字段，意思都是一样的，都是当没有可用资源的时候，生成一个。

这里我们留意到系统的资源池是没有大小限制的，也就是说默认情况下是无上限的，受内存大小限制。

资源的获取和释放对应的方法是Get和Put,也很简洁，返回任意对象interface{}。
~~~
2017/04/17 22:42:43 第0个查询，使用的是ID为2的数据库连接
2017/04/17 22:42:43 第2个查询，使用的是ID为5的数据库连接
2017/04/17 22:42:43 第4个查询，使用的是ID为1的数据库连接
2017/04/17 22:42:44 第3个查询，使用的是ID为4的数据库连接
2017/04/17 22:42:44 第1个查询，使用的是ID为3的数据库连接
~~~
关于系统的资源池，我们需要注意的是它缓存的对象都是临时的，也就说下一次GC的时候，这些存放的对象都会被清除掉。
