在Go语言中，我们很多操作都是通过`go`命令进行的，比如我们要执行go文件的编译，就需要使用`go build`命令，除了`build`命令之外，还有很多常用的命令，这一次我们就统一进行介绍，对常用命令有一个了解，这样我们就可以更容易的开发我们的Go程序了。

## Go 开发工具概览

`go`这个工具，别看名字短小，其实非常强大，是一个强大的开发工具，让我们打开终端，看看这个工具有哪些能力。

```
➜  ~ go                                                                                                          
Go is a tool for managing Go source code.

Usage:

    go command [arguments]

The commands are:

    build       compile packages and dependencies
    clean       remove object files
    doc         show documentation for package or symbol
    env         print Go environment information
    bug         start a bug report
    fix         run go tool fix on packages
    fmt         run gofmt on package sources
    generate    generate Go files by processing source
    get         download and install packages and dependencies
    install     compile and install packages and dependencies
    list        list packages
    run         compile and run Go program
    test        test packages
    tool        run specified go tool
    version     print Go version
    vet         run go tool vet on packages

Use "go help [command]" for more information about a command.

Additional help topics:

    c           calling between Go and C
    buildmode   description of build modes
    filetype    file types
    gopath      GOPATH environment variable
    environment environment variables
    importpath  import path syntax
    packages    description of package lists
    testflag    description of testing flags
    testfunc    description of testing functions

Use "go help [topic]" for more information about that topic.
```

可以发现，go支持的子命令很多，同时还支持查看一些【主题】。我们可以使用`go help [command]`或者`go help [topic]`查看一些命令的使用帮助，或者关于某个主题的信息。大部分go的命令，都是接受一个全路径的包名作为参数，比如我们经常用的`go build`。

## go build

`go build`,是我们非常常用的命令，它可以启动编译，把我们的包和相关的依赖编译成一个可执行的文件。

```
usage: go build [-o output] [-i] [build flags] [packages]
```

`go build`的使用比较简洁，所有的参数都可以忽略，直到只有`go build`，这个时候意味着使用当前目录进行编译，下面的几条命令是等价的：

```
go build

go build .

go build hello.go
```

以上这三种写法，都是使用当前目录编译的意思。因为我们忽略了`packages`,所以自然就使用当前目录进行编译了。从这里我们也可以推测出，`go build`本质上需要的是一个路径，让编译器可以找到哪些需要编译的go文件。`packages`其实是一个相对路径，是相对于我们定义的`GOROOT`和`GOPATH`这两个环境变量的，所以有了`packages`这个参数后，`go build`就可以知道哪些需要编译的go文件了。

```
go build flysnow.org/tools
```

这种方式是指定包的方式，这样会明确的编译我们这个包。当然我们也可以使用通配符。

```
go build flysnow.org/tools/...
```

3个点表示匹配所有字符串，这样`go build`就会编译tools目录下的所有包。

讲到`go build`编译，不能不提跨平台编译，Go提供了编译链工具，可以让我们在任何一个开发平台上，编译出其他平台的可执行文件。

默认情况下，都是根据我们当前的机器生成的可执行文件，比如你的是Linux 64位，就会生成Linux 64位下的可执行文件，比如我的Mac，可以使用go env查看编译环境,以下截取重要的部分。

```
➜  ~ go env
GOARCH="amd64"
GOEXE=""
GOHOSTARCH="amd64"
GOHOSTOS="darwin"
GOOS="darwin"
GOROOT="/usr/local/go"GOTOOLDIR="/usr/local/go/pkg/tool/darwin_amd64"
```

注意里面两个重要的环境变量GOOS和GOARCH,其中GOOS指的是目标操作系统，它的可用值为：

1.  darwin

2.  freebsd

3.  linux

4.  windows

5.  android

6.  dragonfly

7.  netbsd

8.  openbsd

9.  plan9

10.  solaris

一共支持10中操作系统。GOARCH指的是目标处理器的架构，目前支持的有：

1.  arm

2.  arm64

3.  386

4.  amd64

5.  ppc64

6.  ppc64le

7.  mips64

8.  mips64le

9.  s390x

一共支持9中处理器的架构，GOOS和GOARCH组合起来，支持生成的可执行程序种类很多，具体组合参考https://golang.org/doc/install/source#environment。如果我们要生成不同平台架构的可执行程序，只要改变这两个环境变量就可以了，比如要生成linux 64位的程序，命令如下：

```
GOOS=linux GOARCH=amd64 go build flysnow.org/hello
```

前面两个赋值，是更改环境变量，这样的好处是只针对本次运行有效，不会更改我们默认的配置。

以上这些用法差不多够我们用的了，更多关于`go build`的用户可以通过以下命令查看:

```
go help build
```

## go clean

在我们使用`go build`编译的时候，会产生编译生成的文件，尤其是在我们签入代码的时候，并不想把我们生成的文件也签入到我们的Git代码库中，这时候我们可以手动删除生成的文件，但是有时候会忘记，也很麻烦，不小心还是会提交到Git中。要解决这个问题，我们可以使用`go clean`,它可以清理我们编译生成的文件，比如生成的可执行文件，生成obj对象等等。

```
usage: go clean [-i] [-r] [-n] [-x] [build flags] [packages]
```

用法和`go build`基本一样，这样不再进行详细举例演示，可以参考`go build`的使用，更多关于`go clean`的使用，可以使用如下命令查看:

```
go help clean
```

## go run

`go build`是先编译，然后我们在执行可以执行文件来运行我们的程序，需要两步。`go run`这个命令就是可以把这两步合成一步的命令，节省了我们录入的时间，通过`go run`命令，我们可以直接看到输出的结果。

```
➜  ~ go help run
usage: go run [build flags] [-exec xprog] gofiles... [arguments...]

Run compiles and runs the main package comprising the named Go source files.
A Go source file is defined to be a file ending in a literal ".go" suffix.

By default, 'go run' runs the compiled binary directly: 'a.out arguments...'.
If the -exec flag is given, 'go run' invokes the binary using xprog:    'xprog a.out arguments...'.
If the -exec flag is not given, GOOS or GOARCH is different from the system
default, and a program named go_$GOOS_$GOARCH_exec can be found
on the current search path, 'go run' invokes the binary using that program,
for example 'go_nacl_386_exec a.out arguments...'. This allows execution of
cross-compiled programs when a simulator or other execution method is
available.

For more about build flags, see 'go help build'.
```

`go run`命令需要一个go文件作为参数，这个go文件必须包含main包和main函数，这样才可以运行，其他的参数和`go build`差不多。
在运行`go run`的时候，如果需要的话，我们可以给我们的程序传递参数，比如：

```
package mainimport (    "fmt"
    "os")func main() {
    fmt.Println("输入的参数为：",os.Args[1])
}
```

打开终端，输入如下命令执行：

```
go run main.go 12
```

这时候我们就可以看到输出：

```
输入的参数为： 12
```

## go env

在前面讲`go build`的时候，我们使用了`go env`命令查看了我们当前的go环境信息。

```
➜  hello go help env
usage: go env [var ...]

Env prints Go environment information.

By default env prints information as a shell script
(on Windows, a batch file).  If one or more variable
names is given as arguments,  env prints the value of
each named variable on its own line.
```

使用`go env`查看我们的go环境信息，便于我们进行调试，排错等，因为有时候我们会遇到一些莫名其妙的问题，比如本来在MAC上开发，怎么编译出一个Linux的可执行文件等，遇到这类问题时，先查看我们的go环境信息，看看有没有哪里配置错了，一步步排错。

## go install

从其名字上我们不难猜出这个命令是做什么的，它和`go build`类似，不过它可以在编译后，把生成的可执行文件或者库安装到对应的目录下，以供使用。

```
➜  hello go help install
usage: go install [build flags] [packages]

Install compiles and installs the packages named by the import paths,
along with their dependencies.
```

它的用法和`go build`差不多，如果不指定一个包名，就使用当前目录。安装的目录都是约定好的，如果生成的是可执行文件，那么安装在`$GOPATH/bin`目录下；如果是可引用的库，那么安装在`$GOPATH/pkg`目录下。

## go get

`go get`命令，可以从网上下载更新指定的包以及依赖的包，并对它们进行编译和安装。

```
go get github.com/spf13/cobra
```

以上示例，我们就可以从github上直接下载这个go库到我们`GOPATH`工作空间中，以供我们使用。下载的是整个源代码工程，并且会根据它们编译和安装，和执行`go install`类似。

`go get`支持大多数版本控制系统(VCS)，比如我们常用的git，通过它和包依赖管理结合，我们可以在代码中直接导入网络上的包以供我们使用。

如果我们需要更新网络上的一个go工程，加`-u` 标记即可。

```
go get -u github.com/spf13/cobra
```

类似的，启用`-v`标记，可以看到下载的进度以及更多的调试信息。关于`go get` 命令的更多用法，可以使用如下命令查看:

```
go help get
```

## go fmt

这是go提供的最帅的一个命令了，它可以格式化我们的源代码的布局和Go源代码一样的风格，也就是统一代码风格，这样我们再也不用为大括号要不要放到行尾还是另起一行，缩进是使用空格还是tab而争论不休了，都给我们统一了。

```
func main() { 
    fmt.Println("输入的参数为：", os.Args[1]) }
```

比如以上代码，我们执行`go fmt` 格式化后，会变成如下这样：

```
func main() {
    fmt.Println("输入的参数为：", os.Args[1])
}
```

`go fmt`也是接受一个包名作为参数，如果不传递，则使用当前目录。`go fmt`会自动格式化代码文件并保存，它本质上其实是调用的`gofmt -l -w`这个命令，我们看下`gofmt`的使用帮助。

```
➜  hello gofmt -h  
usage: gofmt [flags] [path ...]
  -cpuprofile string
        write cpu profile to this file  -d    display diffs instead of rewriting files  -e    report all errors (not just the first 10 on different lines)  -l    list files whose formatting differs from gofmt's
  -r string
        rewrite rule (e.g., 'a[b:len(a)] -> a[b:]')
  -s    simplify code
  -w    write result to (source) file instead of stdout
```

`go fmt` 为我们统一了代码风格，这样我们在整个团队协作中发现，所有代码都是统一的，像一个人写的一样。所以我们的代码在提交到git库之前，一定要使用`go fmt`进行格式化，现在也有很多编辑器也可以在保存的时候，自动帮我们格式化代码。

## go vet

这个命令不会帮助开发人员写代码，但是它也很有用，因为它会帮助我们检查我们代码中常见的错误。

1.  Printf这类的函数调用时，类型匹配了错误的参数。

2.  定义常用的方法时，方法签名错误。

3.  错误的结构标签。

4.  没有指定字段名的结构字面量。

```
package mainimport (    "fmt")func main() {
    fmt.Printf(" 哈哈",3.14)
}
```

这个例子是一个明显错误的例子，新手经常会犯，这里我们忘记输入了格式化的指令符，这种编辑器是检查不出来的，但是如果我们使用`go vet`就可以帮我们检查出这类常见的小错误。

```
➜  hello go vet
main.go:8: no formatting directive in Printf call
```

看，提示多明显。其使用方式和`go fmt`一样，也是接受一个包名作为参数。

```
usage: go vet [-n] [-x] [build flags] [packages]
```

养成在代码提交或者测试前，使用`go vet`检查代码的好习惯，可以避免一些常见问题。

## go test

该命令用于Go的单元测试，它也是接受一个包名作为参数，如果没有指定，使用当前目录。
`go test`运行的单元测试必须符合go的测试要求。

1.  写有单元测试的文件名，必须以`_test.go`结尾。

2.  测试文件要包含若干个测试函数。

3.  这些测试函数要以Test为前缀，还要接收一个`*testing.T`类型的参数。

```
package mainimport "testing"func TestAdd(t *testing.T) {    if Add(1,2) == 3 {
        t.Log("1+2=3")
    }    if Add(1,1) == 3 {
        t.Error("1+1=3")
    }
}
```

这是一个单元测试，保存在`main_test.go`文件中，对main包里的`Add(a,b int)`函数进行单元测试。
如果要运行这个单元测试，在该文件目录下，执行`go test` 即可。

```
➜  hello go test
PASS
ok      flysnow.org/hello    0.006s
```

以上是打印输出，测试通过。更多关于`go test`命令的使用，请通过如下命令查看。

```
go help test
```

以上这些，主要时介绍的go这个开发工具常用的命令，熟悉了之后可以帮助我们更好的开发编码。《Go语言实战》中针对该部分做了一些介绍，但是比较少，只限于`go build`,`go clean`,`go fmt`,`go vet`这几个命令，这里进行了扩展，还加入了跨平台编译。

其他关于go工具提供的主题介绍，比如package是什么等等，可以直接使用`go help [topic]`命令查看。

```
Additional help topics:

    c           calling between Go and C
    buildmode   description of build modes
    filetype    file types
    gopath      GOPATH environment variable
    environment environment variables
    importpath  import path syntax
    packages    description of package lists
    testflag    description of testing flags
    testfunc    description of testing functions

Use "go help [topic]" for more information about that topic.
```
