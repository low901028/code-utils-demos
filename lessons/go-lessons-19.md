# Go Writer 和 Reader
## 输入和输出

Go Writer 和 Reader接口的设计遵循了Unix的输入和输出，一个程序的输出可以是另外一个程序的输入。他们的功能单一并且纯粹，这样就可以非常容易的编写程序代码，又可以通过组合的概念，让我们的程序做更多的事情。

比如我们在上一篇的Go log日志 中，就介绍了Unix的三种输入输出输入模式，他们对应的Go语言里有专门的实现。

```
var (
    Stdin  = NewFile(uintptr(syscall.Stdin), "/dev/stdin")
    Stdout = NewFile(uintptr(syscall.Stdout), "/dev/stdout")
    Stderr = NewFile(uintptr(syscall.Stderr), "/dev/stderr")
)
```

这三种标准的输入和输出都是一个`*File`，而`*File`恰恰就是同时实现了`io.Writer`和`io.Reader`这两个接口的类型，所以他们同时具备输入和输出的功能，既可以从里面读取数据，又可以往里面写入数据。

Go标准库的io包也是基于Unix这种输入和输出的理念，大部分的接口都是扩展了`io.Writer`和`io.Reader`，大部分的类型也都选择的实现了`io.Writer`和`io.Reader`这两个接口，然后把数据的输入和输出，抽象为流的读写，所以只要实现了这两个接口，都可以使用流的读写功能。

`io.Writer`和`io.Reader`两个接口的高度抽象，让我们不用再面向具体的业务，我们只关注，是读还是写，只要我们定义的方法函数可以接收这两个接口作为参数，那么我们就可以进行流的读写，而不用关心如何读，写到哪里去，这也是面向接口编程的好处。

## Reader和Writer接口

这两个高度抽象的接口，只有一个方法，也体现了Go接口设计的简洁性，只做一件事。

```
// Writer is the interface that wraps the basic Write method.
//
// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
//
// Implementations must not retain p.
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

这是`Wirter`接口的定义，它只有一个`Write`方法，接受一个byte的切片，返回两个值，`n`表示写入的字节数、`err`表示写入时发生的错误。

从其文档注释来看，这个方法是有规范要求的，我们要想实现一个`io.Writer`接口，就要遵循这些规则。

1.  write方法向底层数据流写入len(p)字节的数据，这些数据来自于切片p
2.  返回被写入的字节数n,0 <= n <= len(p)
3.  如果n<len(p), 则必须返回一些非nil的err
4.  如果中途出现问题，也要返回非nil的err
5.  Write方法绝对不能修改切片p以及里面的数据

这些实现`io.Writer`接口的规则，所有实现了该接口的类型都要遵守，不然可能会导致莫名其妙的问题。

```
// Reader is the interface that wraps the basic Read method.
//
// Read reads up to len(p) bytes into p. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered. Even if Read
// returns n < len(p), it may use all of p as scratch space during the call.
// If some data is available but not len(p) bytes, Read conventionally
// returns what is available instead of waiting for more.
//
// When Read encounters an error or end-of-file condition after
// successfully reading n > 0 bytes, it returns the number of
// bytes read. It may return the (non-nil) error from the same call
// or return the error (and n == 0) from a subsequent call.
// An instance of this general case is that a Reader returning
// a non-zero number of bytes at the end of the input stream may
// return either err == EOF or err == nil. The next Read should
// return 0, EOF.
//
// Callers should always process the n > 0 bytes returned before
// considering the error err. Doing so correctly handles I/O errors
// that happen after reading some bytes and also both of the
// allowed EOF behaviors.
//
// Implementations of Read are discouraged from returning a
// zero byte count with a nil error, except when len(p) == 0.
// Callers should treat a return of 0 and nil as indicating that
// nothing happened; in particular it does not indicate EOF.
//
// Implementations must not retain p.
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

这是`io.Reader`接口定义，也只有一个Read方法，这个方法接受一个byte的切片，并返回两个值，一个是读入的字节数，一个是err错误。

从其注释文档看，`io.Reader`接口的规则更多。
1.  Read最多读取len(p)字节的数据，并保存到p。
2.  返回读取的字节数以及任何发生的错误信息
3.  n要满足0 <= n <= len(p)
4.  n<len(p)时，表示读取的数据不足以填满p，这时方法会立即返回，而不是等待更多的数据
5.  读取过程中遇到错误，会返回读取的字节数n以及相应的错误err
6.  在底层输入流结束时，方法会返回n>0的字节，但是err可能时EOF，也可以是nil
7.  在第6种(上面)情况下，再次调用read方法的时候，肯定会返回0,EOF
8.  调用Read方法时，如果n>0时，优先处理处理读入的数据，然后再处理错误err，EOF也要这样处理
9.  Read方法不鼓励返回n=0并且err=nil的情况，

规则稍微比`Write`接口有点多，不过也都比较好理解，注意第8条，即使我们在读取的时候遇到错误，但是也应该也处理已经读到的数据，因为这些已经读到的数据是正确的，如果不进行处理丢失的话，读到的数据就不完整了。

## 示例

对这两个接口了解后，我们就可以尝试使用他们了，现在来看个例子。

```
func main() {    //定义零值Buffer类型变量b
    var b bytes.Buffer    //使用Write方法为写入字符串
    b.Write([]byte("你好"))        //这个是把一个字符串拼接到Buffer里
    fmt.Fprint(&b,",","http://www.flysnow.org")        //把Buffer里的内容打印到终端控制台
    b.WriteTo(os.Stdout)
}
```

这个例子是拼接字符串到`Buffer`里，然后再输出到控制台，这个例子非常简单，但是利用了流的读写，`bytes.Buffer`是一个可变字节的类型，可以让我们很容易的对字节进行操作，比如读写，追加等。`bytes.Buffer`实现了`io.Writer`和`io.Reader`接口，所以我么可以很容易的进行读写操作，而不用关注具体实现。

`b.Write([]byte("你好"))`实现了写入一个字符串，我们把这个字符串转为一个字节切片，然后调用`Write`方法写入，这个就是`bytes.Buffer`为了实现`io.Writer`接口而实现的一个方法，可以帮我们写入数据流。

```
func (b *Buffer) Write(p []byte) (n int, err error) {
    b.lastRead = opInvalid
    m := b.grow(len(p))    return copy(b.buf[m:], p), nil}
```

以上就是`bytes.Buffer`实现`io.Writer`接口的方法，最终我们看到，我们写入的切片会被拷贝到`b.buf`里，这里`b.buf[m:]`拷贝其实就是追加的意思，不会覆盖已经存在的数据。

从实现看，我们发现其实只有`b *Buffer`指针实现了`io.Writer`接口，所以我们示例代码中调用`fmt.Fprint`函数的时候，传递的是一个地址`&b`。

```
func Fprint(w io.Writer, a ...interface{}) (n int, err error) {
    p := newPrinter()
    p.doPrint(a)
    n, err = w.Write(p.buf)
    p.free()    return}
```

这是函数`fmt.Fprint`的实现，它的功能就是为一个把数据`a`写入到一个`io.Writer`接口实现了，具体如何写入，它是不关心的，因为`io.Writer`会做的，它只关心可以写入即可。`w.Write(p.buf)`调用`Wirte`方法写入。

最后的`b.WriteTo(os.Stdout)`是把最终的数据输出到标准的`os.Stdout`里，以便我们查看输出，它接收一个`io.Writer`接口类型的参数，开篇我们讲过`os.Stdout`也实现了这个`io.Writer`接口，所以就可以作为参数传入。

这里我们会发现，很多方法的接收参数都是`io.Writer`接口，当然还有`io.Reader`接口，这就是面向接口的编程，我们不用关注具体实现，只用关注这个接口可以做什么事情，如果我们换成输出到文件里，那么也很容易，只用把`os.File`类型作为参数即可。任何实现了该接口的类型，都可以作为参数。

除了`b.WriteTo`方法外，我们还可以使用`io.Reader`接口的`Read`方法实现数据的读取.

```
    var p [100]byte
    n,err:=b.Read(p[:])
    fmt.Println(n,err,string(p[:n]))
```

这是最原始的方法，使用`Read`方法，n为读取的字节数，然后我们输出打印出来。

因为`byte.Buffer`指针实现了`io.Reader`接口，所以我们还可以使用如下方式读取数据信息。

```
    data,err:=ioutil.ReadAll(&b)
    fmt.Println(string(data),err)
```

`ioutil.ReadAll`接口一个`io.Reader`接口的参数，表明可以从任何实现了`io.Reader`接口的类型里读取全部的数据。

```
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
    buf := bytes.NewBuffer(make([]byte, 0, capacity))    
    // If the buffer overflows, we will get bytes.ErrTooLarge.
    // Return that as an error. Any other panic remains.
    defer func() {
        e := recover()                
        if e == nil {                    
          return
        }               
        if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
                   err = panicErr
        } else {                    
                  panic(e)
        }
    }()
    _, err = buf.ReadFrom(r)    
    return buf.Bytes(), err
}
```

以上是`ioutil.ReadAll`实现的源代码，也非常简单，基本原理是创建一个`byte.Buffer` ,通过这个`byte.Buffer`的`ReadFrom`方法，把`io.Reader`里的数据读取出来，最后通过`byte.Buffer`的`Bytes`方法进行返回最终读取的字节数据信息。

整个流的读取和写入已经被完全抽象啦， `io`包的大部分操作和类型都是基于这两个接口，当然还有http等其他牵涉到数据流、文件流等的，都可以完全用`io.Writer`和`io.Reader`接口来表示，通过这两个接口的连接，我们可以实现任何数据的读写。
