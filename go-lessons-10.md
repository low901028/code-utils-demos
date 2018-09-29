# Go嵌入类型
嵌入类型，或者嵌套类型，这是一种可以把已有的类型声明在新的类型里的一种方式，这种功能对代码复用非常重要。

在其他语言中，有继承可以做同样的事情，但是在Go语言中，没有继承的概念，Go提倡的代码复用的方式是组合，所以这也是嵌入类型的意义所在，组合而不是继承，所以Go才会更灵活。
~~~
type Reader interface {
    Read(p []byte) (n int, err error)
}
type Writer interface {
    Write(p []byte) (n int, err error)
}
type Closer interface {
    Close() error
}
type ReadWriter interface {
    Reader
    Writer
}
type ReadCloser interface {
    Reader
    Closer
}
type WriteCloser interface {
    Writer
    Closer
}
~~~
以上是标准库io包里，我们常用的接口，可以看到ReadWriter接口是嵌入Reader和Reader接口而组合成的新接口，这样我们就不用重复的定义被嵌入接口里的方法，直接通过嵌入就可以了。嵌入类型同样适用于结构体类型，我们再来看个例子：
~~~
type user struct {
    name string
    email string
}
type admin struct {
    user
    level string
}
~~~
嵌入后，被嵌入的类型称之为内部类型、新定义的类型称之为外部类型，这里user就是内部类型，而admin是外部类型。

通过嵌入类型，与内部类型相关联的所有字段、方法、标志符等等所有，都会被外包类型所拥有，就像外部类型自己的一样，这就达到了代码快捷复用组合的目的，而且定义非常简单，只需声明这个类型的名字就可以了。

同时，外部类型还可以添加自己的方法、字段属性等，可以很方便的扩展外部类型的功能。
~~~
func main() {
    ad:=admin{user{"张三","zhangsan@flysnow.org"},"管理员"}
    fmt.Println("可以直接调用,名字为：",ad.name)
    fmt.Println("也可以通过内部类型调用,名字为：",ad.user.name)
    fmt.Println("但是新增加的属性只能直接调用，级别为：",ad.level)
}
~~~
以上是嵌入类型的使用，可以看到，我们在初始化的时候，采用的是字面值的方式，所以要按其定义的结构进行初始化，先初始化user这个内部类型的，再初始化新增的level 属性。

对于内部类型的属性和方法访问上，我们可以用外部类型直接访问，也可以通过内部类型进行访问；但是我们为外部类型新增的方法属性字段，只能使用外部类型访问，因为内部类型没有这些。

当然，外部类型也可以声明同名的字段或者方法，来覆盖内部类型的，这种情况方法比较多，我们以方法为例
~~~
func main() {
    ad:=admin{user{"张三","zhangsan@flysnow.org"},"管理员"}
    ad.user.sayHello()
    ad.sayHello()
}
type user struct {
    name string
    email string
}
type admin struct {
    user
    level string
}
func (u user) sayHello(){
    fmt.Println("Hello，i am a user")
}
func (a admin) sayHello(){
    fmt.Println("Hello，i am a admin")
}
~~~
内部类型user有一个sayHello方法，外部类型对其进行了覆盖，同名重写sayHello，然后我们在main方法里分别访问这两个类型的方法，打印输出:
~~~
Hello，i am a user
Hello，i am a admin
~~~
从输出中看，方法sayHello被成功覆盖了。

嵌入类型的强大，还体现在：如果内部类型实现了某个接口，那么外部类型也被认为实现了这个接口。我们稍微改造下例子看下。
~~~
func main() {
    ad:=admin{user{"张三","zhangsan@flysnow.org"},"管理员"}
    sayHello(ad.user)//使用user作为参数
    sayHello(ad)//使用admin作为参数
}
type Hello interface {
    hello()
}
func (u user) hello(){
    fmt.Println("Hello，i am a user")
}
func sayHello(h Hello){
    h.hello()
}
~~~
这个例子原来的结构体类型user和admin的定义不变，新增了一个接口Hello,然后让user类型实现这个接口，最后我们定义了一个sayHello方法，它接受一个Hello接口类型的参数，最终我们在main函数演示的时候，发现不管是user类型，还是admin类型作为参数传递给sayHello方法的时候，都可以正常调用。

这里就可以说明admin实现了接口Hello,但是我们又没有显示的声明类型admin实现，所以这个实现是通过内部类型user实现的，因为admin包含了user所有的方法函数，所以也就实现了接口Hello。

当然外部类型也可以重新实现，只需要像上面例子一样覆盖同名的方法即可。这里要说明的是，不管我们如何同名覆盖，都不会影响内部类型，我们还可以通过访问内部类型来访问它的方法、属性字段等。

嵌入类型的定义，是Go为了方便我们扩展或者修改已有类型的行为，是为了宣传组合这个概念而设计的，所以我们经常使用组合，灵活运用组合，扩展出更多的我们需要的类型结构。
