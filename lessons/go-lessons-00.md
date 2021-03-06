本篇特意针对Go语言的开发环境搭建、配置、编辑器选型、不同平台程序生成等做了详细的介绍。

### 下载
要搭建Go语言开发环境，我们第一步要下载go的开发工具包，下载最新稳定版本。Go为我们所熟知的所有平台架构提供了开发工具包，比如我们熟知的Linux、Mac和Windows，其他的还有FreeBSD等。

我们可以根据自己的机器操作系统选择相应的开发工具包，比如你的是Windows 64位的，就选择windows-amd64的工具包；是Linux 32位的就选择linux-386的工具包。可以自己查看下自己的操作系统，然后选择，Mac的现在都是64位的，直接选择就可以了。

开发工具包又分为安装版和压缩版。安装版是Mac和Windows特有的，他们的名字类似于：
go1.7.4.darwin-amd64.pkg
go1.7.4.windows-386.msi
go1.7.4.windows-amd64.msi
安装版，顾名思义，双击打开会出现安装向导，让你选择安装的路径，帮你设置好环境比安康等信息，比较省事方便一些。

压缩版的就是一个压缩文件，可以解压得到里面的内容，他们的名字类似于：
go1.7.4.darwin-amd64.tar.gz
go1.7.4.linux-386.tar.gz
go1.7.4.linux-amd64.tar.gz
go1.7.4.windows-386.zip
go1.7.4.windows-amd64.zip
压缩版我们下载后需要解压，然后自己移动到要存放的路径下，并且配置环境变量等信息，相比安装版来说，比较复杂一些，手动配置的比较多。

根据自己的操作系统选择后，就可以下载开发工具包了，Go语言的官方下载地址是 https://golang.org/dl/ 可以打开选择版本下载，如果该页面打不开，或者打开了下载不了，可以使用镜像网站 http://mirrors.flysnow.org/ ,打开后搜索或者找到Golang，选择相应的版本下载，这个镜像网站会同步更新官方版本，基本上都是最新版，可以放心使用。

#### Linux下安装
我们以Ubuntu 64位为例进行演示，CentOS等其他Linux发行版大同小异。
下载go1.7.4.linux-amd64.tar.gz后，进行解压，你可以采用自带的解压软件解压，
如果没有可以在终端行使用tar命令行工具解压，我们这里选择的安装目录是/usr/local/go,可以使用如下命令：
~~~
tar -C /usr/local -xzf go1.7.4.linux-amd64.tar.gz
~~~
如果提示没有权限，在最前面加上sudo以root用户的身份运行。运行后，在／usr/local/下就可以看到go目录了。如果是自己用软件解压的，可以拷贝到/usr/local/go下，但是要保证你的go文件夹下是bin、src、doc等目录，不要go文件夹下又是一个go文件夹，这样就双重嵌套了。

然后就要配置环境变量了，Linux下又两个文件可以配置，其中/etc/profile是针对所有用户都有效的；$HOME/.profile是针对当前用户有效的，可以根据自己的情况选择。

针对所有用户的需要重启电脑才可以生效；针对当前用户的，在终端里使用source命令加载这个$HOME/.profile即可生效。
~~~
source ~/.profile
~~~
使用文本编辑器比如VIM编辑他们中的任意一个文件，在文件的末尾添加如下配置保存即可：
~~~
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin
~~~
其中GOROOT环境变量表示我们GO的安装目录，这样其他软件比如我们使用的Go开发IDE就可以自动的找到我们的Go安装目录，达到自动配置Go SDK的目的。

第二句配置是把/usr/local/go/bin这个目录加入到环境变量PATH里，这样我可以在终端里直接输入go等常用命令使用了，而不用再加上/usr/local/go/bin这一串绝对路径，更简洁方便。

以上配置好之后，我们打开终端，属于如下命令，就可以看到go的版本等信息了。
~~~
 ~ go version
go version go1.7.4 linux/amd64
~~~
这就说明我们已经安装go成功了，如果提示go这个命令找不到，说明我们配置还不对，主要在PATH这个环境变量，仔细检查，直到可以正常输出为止。

#### Mac下安装
Mac分为压缩版和安装版，他们都是64位的。压缩版和Linux的大同小异，因为Mac和Linux都是基于Unix，终端这一块基本上是相同的。

压缩版解压后，就可以和Linux一样放到一个目录下，这里也以/usr/local/go/为例。在配置环境变量的时候，针对所有用户和Linux是一样的，都是/etc/profile这个文件；针对当前用户，Mac下是$HOME/.bash_profile，其他配置都一样，包括编辑sudo权限和生效方式，最后在终端里测试：
~~~
~ go version
go version go1.7.4 darwin/amd64
~~~

Mac安装版下载后双击可以看到安装界面，按照提示一步步选择操作即可。安装版默认安装目录是/usr/local/go，并且也会自动的把/usr/local/go/bin目录加入到PATH环境变量中，重新打开一个终端，就可以使用go version进行测试了，更快捷方便一些。

#### Windows下安装
Windows也有压缩版和安装版，又分为32和64位以供选择，不过目前大家都是64位，选择这个更好一些。

Window的压缩版是一个ZIP压缩包，下载后使用winrar等软件就可以解压，解压后要选择一个存放目录，比如c:\Go下，这个c:\Go就是Go的安装目录了，他里面有bin、src、doc等目录。

然后就是环境变量的配置，Window也和Linux一样分为针对所有用户的系统变量，和针对当前用户的用户变量设置，可以自行选择，比如系统变量，针对所有用户都有效。

以Window 7为例，右击我的电脑->属性会打开系统控制面板，然后在左侧找到高级系统设置点击打开，会在弹出的界面最下方看到环境变量按钮，点击它，就可以看到环境变量配置界面了。上半部分是用户变量配置，下半部分是系统变量配置。

我们在系统变量里点击新建，变量名输入GOROOT，变量值是我们刚刚安装的go路径c:\Go,这样就配置好了GO目录的安装路径了。

然后修改PATH系统变量，在变量值里添加%%GOROOT\bin路径，和其他PATH变量以;(分号，Linux下是冒号)分割即可。这样我们就可以在CMD里直接输入go命令使用了。

打开我们的终端，输入go version测试下，好了的话就可以看到输出的信息了。

Window的安装版相比来说就比较简单一些，双击就可以按照提示一步步安装，默认安装路径是c:\Go,并且会配置好PATH环境变量，可以直接打开CMD终端使用。

### 设置工作目录【很多没有理解这个地方 导致在后续的使用中出现意想不到的bug <_>】

工作目录就是我们用来存放开发的源代码的地方，对应的也是Go里的GOPATH这个环境变量。这个环境变量指定之后，我们编译源代码等生成的文件都会放到这个目录下，GOPATH环境变量的配置参考上面的安装Go，配置到/etc/profile或者Windows下的系统变量里。

这个工作目录我们可以根据自己的设置指定，比如我的Mac在$HOME/code/go下，Window的可以放到d:\code\go下等。该目录下有3个子目录，他们分别是：
.
├── bin
├── pkg
└── src
- bin文件夹存放go install命名生成的可执行文件，可以把GOPATH/bin路径加入到PATH环境变量里，就和我们上面配置的$GOROOT/bin一样，这样就可以直接在终端里使用我们go开发生成的程序了。
- pkg文件夹是存在go编译生成的文件。
- src存放的是我们的go源代码，不同工程项目的代码以包名区分。

#### go项目工程结构

配置好工作目录后，就可以编码开发了，在这之前，我们看下go的通用项目结构,这里的结构主要是源代码相应地资源文件存放目录结构。

我们知道源代码都是存放在GOPATH的src目录下，那么多个多个项目的时候，怎么区分呢？答案是通过包，使用包来组织我们的项目目录结构。有过java开发的都知道，使用包进行组织代码，包以网站域名开头就不会有重复，比如我的个人网站是flysnow.org,我就可以以·flysnow.org·的名字创建一个文件夹，我自己的go项目都放在这个文件夹里，这样就不会和其他人的项目冲突，包名也是唯一的。

如果没有个人域名，现在流行的做法是使用你个人的github.com，因为每个人的是唯一的，所以也不会有重复。
src
├── github.com
├── golang.org
├── gopkg.in
├── qiniupkg.com
└── sourcegraph.com
如上，src目录下跟着一个个域名命名的文件夹。再以github.com文件夹为例，它里面又是以github用户名命名的文件夹，用于存储属于这个github用户编写的go源代码。
src/github.com/spf13
├── afero
├── cast
├── cobra
├── fsync
├── hugo
├── jwalterweatherman
├── nitro
├── pflag
└── viper
那么我们如何引用一个包呢，也就是go里面的import。其实非常简单，通过包路径，包路径就是从src目录开始，逐级文件夹的名字用/连起来就是我们需要的包名，比如：
~~~
import (
    "github.com/spf13/hugo/commands"
)
~~~
都准备好了，让我们创建一个hello项目，测试一下。我的项目的路径为src/flysnow.org/hello/。
~~~
package main
import ( 
   "fmt"
)
func main() {
    fmt.Println("Hello World")
}
~~~
Go版Hello World非常简单。在src/flysnow.org/hello/目录下运行go run main.go命令就可以看到打印的输出Hello World，下面解释下这段代码。

- package 是一个关键字，定义一个包，和Java里的package一样，也是模块化的关键。
- main包是一个特殊的包名，它表示当前是一个可执行程序，而不是一个库。
- import 也是一个关键字，表示要引入的包，和Java的import关键字一样，引入后才可以使用它。
- fmt是一个包名，这里表示要引入fmt这个包，这样我们就可以使用它的函数了。
- main函数是主函数，表示程序执行的入口，Java也有同名函数，但是多了一个String[]类型的参数。
- Println是fmt包里的函数，和Java里的system.out.println作用类似，这里输出一段文字。
整段代码非常简洁，关键字、函数、包等和Java非常相似，不过注意，go是不需要以;(分号)结尾的。

### 安装程序
安装的意思，就是生成可执行的程序，以供我们使用，为此go为我们提供了很方便的install命令，可以快速的把我们的程序安装到$GOAPTH/bin目录下。
~~~
go install flysnow.org/hello
~~~
打开终端，运行上面的命令即可，install后跟全路径的包名。 然后我们在终端里运行hello就看到打印的Hello World了。
~~~
~ hello
Hell World
~~~
跨平台编译

以前运行和安装，都是默认根据我们当前的机器生成的可执行文件，比如你的是Linux 64位，就会生成Linux 64位下的可执行文件，比如我的Mac，可以使用go env查看编译环境,以下截取重要的部分。
~~~
~ go env
GOARCH="amd64"
GOEXE=""
GOHOSTARCH="amd64"
GOHOSTOS="darwin"
GOOS="darwin"
GOROOT="/usr/local/go"
GOTOOLDIR="/usr/local/go/pkg/tool/darwin_amd64"
~~~
注意里面两个重要的环境变量GOOS和GOARCH,其中GOOS指的是目标操作系统，它的可用值为：
~~~
darwin
freebsd
linux
windows
android
dragonfly
netbsd
openbsd
plan9
solaris
~~~
一共支持10中操作系统。GOARCH指的是目标处理器的架构，目前支持的有：
~~~
arm
arm64
386
amd64
ppc64
ppc64le
mips64
mips64le
s390x
~~~
一共支持9中处理器的架构，GOOS和GOARCH组合起来，支持生成的可执行程序种类很多，具体组合参考https://golang.org/doc/install/source#environment。如果我们要生成不同平台架构的可执行程序，只要改变这两个环境变量就可以了，比如要生成linux 64位的程序，命令如下：
~~~
GOOS=linux GOARCH=amd64 go build flysnow.org/hello
~~~
前面两个赋值，是更改环境变量，这样的好处是只针对本次运行有效，不会更改我们默认的配置。

### 获取远程包（go get）
go提供了一个获取远程包的工具go get,他需要一个完整的包名作为参数，只要这个完成的包名是可访问的，就可以被获取到，比如我们获取一个CLI的开源库：
~~~
go get -v github.com/spf13/cobra/cobra
~~~
就可以下载这个库到我们$GOPATH/src目录下了，这样我们就可以像导入其他包一样import了。

特别提醒，go get的本质是使用源代码控制工具下载这些库的源代码，比如git，hg等，所以在使用之前必须确保安装了这些源代码版本控制工具。

### Go编辑器推荐
Go采用的是UTF-8的文本文件存放源代码，所以原则上你可以使用任何一款文本编辑器，这里推荐几款比较流行的。

对于新手来说，我推荐功能强大的IDE，功能强大，使用方便，比如jetbrains idea+golang插件，上手容易，而且它家的IDE都一样，会一个都会了，包括菜单、快捷键等。值得高兴的是jetbrains针对Go这门语言推出了专用IDE gogland，也足以证明go的流行以及jetbrains的重视。goglang地址为 https://www.jetbrains.com/go/,可以前往下载使用。

其次可以推荐微软的VS Code以及Sublime Text，这两款编辑器插件强大，快捷键方便，都对Go支持的很好，也拥有大量的粉丝。

最后推荐老牌的VIM，这个不用多介绍，大家都知道。

到这里，整个Go开发环境就详细介绍完了，不光有环境安装搭建，还有目录结构、常用命令使用等都进行了介绍，这篇文章看完后，已经入门了Go了，剩下的再看看Go的语法和库，就可以很流畅的编写Go程序了。
