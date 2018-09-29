## 什么是单元测试

相信我们做程序员的，对单元测试都不陌生。单元测试一般是用来测试我们的代码逻辑有没有问题，有没有按照我们期望的运行，以保证代码质量。

大多数的单元测试，都是对某一个函数方法进行测试，以尽可能的保证没有问题或者问题可被我们预知。为了达到这个目的，我们可以使用各种手段、逻辑，模拟不同的场景进行测试。

这里我们在`package main`里定义一个函数`Add`，求两个数之和的函数，然后我们使用单元测试进行求和逻辑测试。

*main.go*

```
func Add(a,b int) int{    return a+b
}
```

*main_test.go*

```
func TestAdd(t *testing.T) {
    sum := Add(1,2)    if sum == 3 {
        t.Log("the result is ok")
    } else {
        t.Fatal("the result is wrong")
    }
}
```

然后我们在终端的项目目录下运行`go test -v`就可以看到测试结果了。

```
➜  hello go test -v
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
        main_test.go:26: the result is ok
PASS
ok      flysnow.org/hello       0.007s
```

有测试成功PASS标记，并且打印出我们想要的结果。更多关于`go test`的用法参考以前写的*Go开发工具* http://www.flysnow.org/2017/03/08/go-in-action-go-tools.html,这里就不要细说了。

Go语言为我们提供了测试框架，以便帮助我们更容易的进行单元测试，但是要使用这个框架，需要遵循如下几点规则：

1.  含有单元测试代码的go文件必须以`_test.go`结尾，Go语言测试工具只认符合这个规则的文件

2.  单元测试文件名`_test.go`前面的部分最好是被测试的方法所在go文件的文件名，比如例子中是`main_test.go`，因为测试的`Add`函数，在`main.go`文件里

3.  单元测试的函数名必须以`Test`开头，是可导出公开的函数

4.  测试函数的签名必须接收一个指向`testing.T`类型的指针，并且不能返回任何值

5.  函数名最好是Test+要测试的方法函数名，比如例子中是`TestAdd`，表示测试的是`Add`这个这个函数

遵循以上规则，我们就可以很容易的编写单元测试了，单元测试的重点在于测试代码的逻辑，场景等，以便尽可能的测试全面，保障代码质量逻辑。

## 表组测试

还有一种单元测试方法叫表组测试，这个和基本的单元测试非常相似，只不过它是有好几个不同的输入以及输出组成的一组单元测试。

比如上个例子中，我们测试了`1+2`，如果我们再加上`3+4`,`9+2`等，这就有了好几个输入，同时对应的也有好几个输出，这种一次性测试很多个输入输出场景的测试，就是表组测试。

```
func TestAdd(t *testing.T) {
    sum := Add(1,2)    if sum == 3 {
        t.Log("the result is ok")
    } else {
        t.Fatal("the result is wrong")
    }

    sum=Add(3,4)        if sum == 7 {
        t.Log("the result is ok")
    } else {
        t.Fatal("the result is wrong")
    }
}
```

## 模拟调用

单元测试的原则，就是你所测试的函数方法，不要受到所依赖环境的影响，比如网络访问等，因为有时候我们运行单元测试的时候，并没有联网，那么总不能让单元测试因为这个失败吧？所以这时候模拟网络访问就有必要了。

针对模拟网络访问，标准库了提供了一个httptest包，可以让我们模拟http的网络调用，下面举个例子了解使用。

首先我们创建一个处理HTTP请求的函数，并注册路由

```
package commonimport (    "net/http"
    "encoding/json")func Routes(){
    http.HandleFunc("/sendjson",SendJSON)
}func SendJSON(rw http.ResponseWriter,r *http.Request){
    u := struct {
        Name string
    }{
        Name:"张三",
    }

    rw.Header().Set("Content-Type","application/json")
    rw.WriteHeader(http.StatusOK)
    json.NewEncoder(rw).Encode(u)
}
```

非常简单，这里是一个`/sendjson`API，当我们访问这个API时，会返回一个JSON字符串。现在我们对这个API服务进行测试，但是我们又不能时时刻刻都启动着服务，所以这里就用到了外部终端对API的网络访问请求。

```
func init()  {
    common.Routes()
}func TestSendJSON(t *testing.T){
    req,err:=http.NewRequest(http.MethodGet,"/sendjson",nil)    if err!=nil {
        t.Fatal("创建Request失败")
    }

    rw:=httptest.NewRecorder()
    http.DefaultServeMux.ServeHTTP(rw,req)

    log.Println("code:",rw.Code)

    log.Println("body:",rw.Body.String())
}
```

运行这个单元测试，就可以看到我们访问`/sendjson`API的结果里，并且我们没有启动任何HTTP服务就达到了目的。这个主要利用`httptest.NewRecorder()`创建一个`http.ResponseWriter`，模拟了真实服务端的响应，这种响应时通过调用`http.DefaultServeMux.ServeHTTP`方法触发的。

还有一个模拟调用的方式，是真的在测试机上模拟一个服务器，然后进行调用测试。

```
func mockServer() *httptest.Server {    //API调用处理函数
    sendJson := func(rw http.ResponseWriter, r *http.Request) {
        u := struct {
            Name string
        }{
            Name: "张三",
        }

        rw.Header().Set("Content-Type", "application/json")
        rw.WriteHeader(http.StatusOK)
        json.NewEncoder(rw).Encode(u)
    }        //适配器转换
    return httptest.NewServer(http.HandlerFunc(sendJson))
}func TestSendJSON(t *testing.T) {    //创建一个模拟的服务器
    server := mockServer()        defer server.Close()        //Get请求发往模拟服务器的地址
    resq, err := http.Get(server.URL)        if err != nil {
        t.Fatal("创建Get失败")
    }    defer resq.Body.Close()

    log.Println("code:", resq.StatusCode)
    json, err := ioutil.ReadAll(resq.Body)        if err != nil {
        log.Fatal(err)
    }
    log.Printf("body:%s\n", json)
}
```

模拟服务器的创建使用的是`httptest.NewServer`函数，它接收一个`http.Handler`处理API请求的接口。
代码示例中使用了Hander的适配器模式，`http.HandlerFunc`是一个函数类型，实现了`http.Handler`接口，这里是强制类型转换，不是函数的调用.

这个创建的模拟服务器，监听的是本机IP`127.0.0.1`，端口是随机的。接着我们发送Get请求的时候，不再发往`/sendjson`，而是模拟服务器的地址`server.URL`，剩下的就和访问正常的URL一样了，打印出结果即可。

## 测试覆盖率

我们尽可能的模拟更多的场景来测试我们代码的不同情况，但是有时候的确也有忘记测试的代码，这时候我们就需要测试覆盖率作为参考了。

由单元测试的代码，触发运行到的被测试代码的代码行数占所有代码行数的比例，被称为测试覆盖率，代码覆盖率不一定完全精准，但是可以作为参考，可以帮我们测量和我们预计的覆盖率之间的差距，`go test`工具，就为我们提供了这么一个度量测试覆盖率的能力。

*main.go*

```
func Tag(tag int){    switch tag {        case 1:
        fmt.Println("Android")        case 2:
        fmt.Println("Go")        case 3:
        fmt.Println("Java")        default:
        fmt.Println("C")

    }
}
```

*main_test.go*

```
func TestTag(t *testing.T) {
    Tag(1)
    Tag(2)

}
```

现在我们使用`go test`工具运行单元测试，和前几次不一样的是，我们要显示测试覆盖率，所以要多加一个参数`-coverprofile`,所以完整的命令为：`go test -v -coverprofile=c.out`，`-coverprofile`是指定生成的覆盖率文件，例子中是`c.out`，这个文件一会我们会用到。现在我们看终端输出，已经有了一个覆盖率。

```
=== RUN   TestTag
Android
Go
--- PASS: TestTag (0.00s)
PASS
coverage: 60.0% of statements
ok      flysnow.org/hello       0.005s
```

`coverage: 60.0% of statements`，60%的测试覆盖率，还没有到100%，那么我们看看还有那些代码没有被测试到。这就需要我们刚刚生成的测试覆盖率文件`c.out`生成测试覆盖率报告了。生成报告有go为我们提供的工具，使用`go tool cover -html=c.out -o=tag.html`，即可生成一个名字为`tag.html`的HTML格式的测试覆盖率报告，这里有详细的信息告诉我们哪一行代码测试到了，哪一行代码没有测试到。

![image](http://upload-images.jianshu.io/upload_images/5525735-bce494073c8a347f?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

从上图中可以看到，标记为绿色的代码行已经被测试了；标记为红色的还没有测试到，有2行的，现在我们根据没有测试到的代码逻辑，完善我的单元测试代码即可。

```
func TestTag(t *testing.T) {
    Tag(1)
    Tag(2)
    Tag(3)
    Tag(6)

}
```

单元测试完善为如上代码，再运行单元测试，就可以看到测试覆盖率已经是100%了，大功告成。
