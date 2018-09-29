# GO接口
接口是一种约定，它是一个抽象的类型，和我们见到的具体的类型如int、map、slice等不一样。具体的类型，我们可以知道它是什么，并且可以知道可以用它做什么；但是接口不一样，接口是抽象的，它只有一组接口方法，我们并不知道它的内部实现，所以我们不知道接口是什么，但是我们知道可以利用它提供的方法做什么。

抽象就是接口的优势，它不用和具体的实现细节绑定在一起，我们只需定义接口，告诉编码人员它可以做什么，这样我们可以把具体实现分开，这样编码就会更加灵活方面，适应能力也会非常强。
~~~
func main() {
    var b bytes.Buffer
    fmt.Fprint(&b,"Hello World")
    fmt.Println(b.String())
}
~~~
以上就是一个使用接口的例子，我们先看下fmt.Fprint函数的实现。
~~~
func Fprint(w io.Writer, a ...interface{}) (n int, err error) {
    p := newPrinter()
    p.doPrint(a)
    n, err = w.Write(p.buf)
    p.free()
    return
}
~~~
从上面的源代码中，我们可以看到，fmt.Fprint函数的第一个参数是io.Writer这个接口，所以只要实现了这个接口的具体类型都可以作为参数传递给fmt.Fprint函数，而bytes.Buffer恰恰实现了io.Writer接口，所以可以作为参数传递给fmt.Fprint函数。

### 内部实现

我们前面提过接口是用来定义行为的类型，它是抽象的，这些定义的行为不是由接口直接实现，而是通过方法由用户定义的类型实现。如果用户定义的类型，实现了接口类型声明的所有方法，那么这个用户定义的类型就实现了这个接口，所以这个用户定义类型的值就可以赋值给接口类型的值。
~~~
func main() {
    var b bytes.Buffer
    fmt.Fprint(&b, "Hello World")
    var w io.Writer
    w = &b
    fmt.Println(w)
}
~~~
这里例子中，因为bytes.Buffer实现了接口io.Writer,所以我们可以通过w = &b赋值，这个赋值的操作会把定义类型的值存入接口类型的值。

赋值操作执行后，如果我们对接口方法执行调用，其实是调用存储的用户定义类型的对应方法，这里我们可以把用户定义的类型称之为实体类型。

我们可以定义很多类型，让它们实现一个接口，那么这些类型都可以赋值给这个接口，这时候接口方法的调用，其实就是对应实体类型对应方法的调用，这就是多态。
~~~
func main() {
    var a animal
    var c cat
    
    a=c
    a.printInfo()    //使用另外一个类型赋值

    var d dog
    a=d
    a.printInfo()
}
type animal interface {
    printInfo()
}

type cat inttype dog int

func (c cat) printInfo(){
    fmt.Println("a cat")
}
func (d dog) printInfo(){
    fmt.Println("a dog")
}
~~~
以上例子演示了一个多态。我们定义了一个接口animal,然后定义了两种类型cat和dog实现了接口animal。在使用的时候，分别把类型cat的值c、类型dog的值d赋值给接口animal的值a,然后分别执行a的printInfo方法，可以看到不同的输出。
~~~
output：
a cat
a dog
~~~
我们看下接口的值被赋值后，接口值内部的布局。接口的值是一个两个字长度的数据结构，第一个字包含一个指向内部表结构的指针，这个内部表里存储的有实体类型的信息以及相关联的方法集；第二个字包含的是一个指向存储的实体类型值的指针。所以接口的值结构其实是两个指针，这也可以说明接口其实一个引用类型。

### 方法集
我们都知道，如果要实现一个接口，必须实现这个接口提供的所有方法，但是实现方法的时候，我们可以使用指针接收者实现，也可以使用值接收者实现，这两者是有区别的，下面我们就好好分析下这两者的区别。
~~~
func main() {
    var c cat    //值作为参数传递
    invoke(c)
}
//需要一个animal接口作为参数
func invoke(a animal){
    a.printInfo()
}
type animal interface {
    printInfo()
}
type cat int

//值接收者实现animal接口
func (c cat) printInfo(){
    fmt.Println("a cat")
}
~~~
还是原来的例子改改，增加一个invoke函数，该函数接收一个animal接口类型的参数，例子中传递参数的时候，也是以类型cat的值c传递的，运行程序可以正常执行。现在我们稍微改造一下，使用类型cat的指针&c作为参数传递。
~~~
func main() {
    var c cat    
    //指针作为参数传递
    invoke(&c)
}
~~~
只修改这一处，其他保持不变，我们运行程序，发现也可以正常执行。通过这个例子我们可以得出结论：实体类型以值接收者实现接口的时候，不管是实体类型的值，还是实体类型值的指针，都实现了该接口。

下面我们把接收者改为指针试试。
~~~
func main() {
    var c cat
    //值作为参数传递
    invoke(c)
}
//需要一个animal接口作为参数
func invoke(a animal){
    a.printInfo()
}
type animal interface {
    printInfo()
}
type cat int

//指针接收者实现animal接口
func (c *cat) printInfo(){
    fmt.Println("a cat")
}
~~~
这个例子中把实现接口的接收者改为指针，但是传递参数的时候，我们还是按值进行传递，点击运行程序，会出现以下异常提示：
~~~
./main.go:10: cannot use c (type cat) as type animal in argument to invoke:
    cat does not implement animal (printInfo method has pointer receiver)
~~~
提示中已经很明显的告诉我们，说cat没有实现animal接口，因为printInfo方法有一个指针接收者，所以cat类型的值c不能作为接口类型animal传参使用。下面我们再稍微修改下，改为以指针作为参数传递。
~~~
func main() {
    var c cat    
    //指针作为参数传递
    invoke(&c)
}
~~~
其他都不变，只是把以前使用值的参数，改为使用指针作为参数，我们再运行程序，就可以正常运行了。由此可见实体类型以指针接收者实现接口的时候，只有指向这个类型的指针才被认为实现了该接口

现在我们总结下这两种规则，首先以方法接收者是值还是指针的角度看。
![image.png](https://upload-images.jianshu.io/upload_images/5525735-a07d1b38cda6f0d4.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

上面的表格可以解读为：如果是值接收者，实体类型的值和指针都可以实现对应的接口；如果是指针接收者，那么只有类型的指针能够实现对应的接口。

其次我们我们以实体类型是值还是指针的角度看。
![image.png](https://upload-images.jianshu.io/upload_images/5525735-8ce3c555bfc6427a.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

上面的表格可以解读为：类型的值只能实现值接收者的接口；指向类型的指针，既可以实现值接收者的接口，也可以实现指针接收者的接口。
