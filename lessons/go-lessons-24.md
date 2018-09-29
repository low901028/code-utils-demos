# Go反射
和Java语言一样，Go也实现运行时反射，这为我们提供一种可以在运行时操作任意类型对象的能力。比如我们可以查看一个接口变量的具体类型，看看一个结构体有多少字段，如何修改某个字段的值等等。

### TypeOf和ValueOf

在Go的反射定义中，任何接口都会由两部分组成的，一个是接口的具体类型，一个是具体类型对应的值。比如var i int = 3 ，因为interface{}可以表示任何类型，所以变量i可以转为interface{}，所以可以把变量i当成一个接口，那么这个变量在Go反射中的表示就是<Value,Type>，其中Value为变量的值3,Type变量的为类型int。

在Go反射中，标准库为我们提供两种类型来分别表示他们reflect.Value和reflect.Type，并且提供了两个函数来获取任意对象的Value和Type。
~~~
func main() {
    u:= User{"张三",20}
    t:=reflect.TypeOf(u)
    fmt.Println(t)
}
type User struct{
    Name string
    Age int
}
~~~
- reflect.TypeOf可以获取任意对象的具体类型，这里通过打印输出可以看到是main.User这个结构体类型。
- reflect.TypeOf函数接受一个空接口interface{}作为参数，所以这个方法可以接受任何类型的对象。

接着上面的例子，我们看下如何反射获取一个对象的Value。
~~~
    v:=reflect.ValueOf(u)
    fmt.Println(v)
~~~
和TypeOf函数一样，也可以接受任意对象，可以看到打印输出为{张三 20}。对于以上这两种输出，Go语言还通过fmt.Printf函数为我们提供了简便的方法。
~~~
    fmt.Printf("%T\n",u)
    fmt.Printf("%v\n",u)
~~~
这个例子和以上的例子中的输出一样。

### reflect.Value转原始类型

上面的例子我们可以通过reflect.ValueOf函数把任意类型的对象转为一个reflect.Value，那我们如果我们想逆向转过回来呢，其实也是可以的，reflect.Value为我们提供了Inteface方法来帮我们做这个事情。继续接上面的例子：
~~~
    u1:=v.Interface().(User)
    fmt.Println(u1)
~~~
这样我们就又还原为原来的User对象了,通过打印的输出就可以验证。这里可以还原的原因是因为在Go的反射中，把任意一个对象分为reflect.Value和reflect.Type，而reflect.Value又同时持有一个对象的reflect.Value和reflect.Type,所以我们可以通过reflect.Value的Interface方法实现还原。现在我们看看如何从一个reflect.Value获取对应的reflect.Type。
~~~
    t1:=v.Type()
    fmt.Println(t1)
~~~
如上例中，通过reflect.Value的Type方法就可以获得对应的reflect.Type。

### 获取类型底层类型
底层的类型是什么意思呢？其实对应的主要是基础类型，接口、结构体、指针这些，因为我们可以通过type关键字声明很多新的类型，比如上面的例子，对象u的实际类型是User，但是对应的底层类型是struct这个结构体类型，我们来验证下。
~~~
fmt.Println(t.Kind())
~~~
通过Kind方法即可获取，非常简单，当然我们也可以使用Value对象的Kind方法，他们是等价的。

Go语言提供了以下这些最底层的类型，可以看到，都是最基本的。
~~~
const (
    Invalid Kind = iota
    Bool
    Int
    Int8
    Int16
    Int32
    Int64
    Uint
    Uint8
    Uint16
    Uint32
    Uint64
    Uintptr
    Float32
    Float64
    Complex64
    Complex128
    Array
    Chan
    Func
    Interface
    Map
    Ptr
    Slice
    String
    Struct
    UnsafePointer
)
~~~

### 遍历字段和方法
通过反射，我们可以获取一个结构体类型的字段,也可以获取一个类型的导出方法，这样我们就可以在运行时了解一个类型的结构，这是一个非常强大的功能。
~~~
    for i:=0;i<t.NumField();i++ {
        fmt.Println(t.Field(i).Name)
    }    
    for i:=0;i<t.NumMethod() ;i++  {
        fmt.Println(t.Method(i).Name)
    }
~~~
这个例子打印出结构体的所有字段名以及该结构体的方法。NumField方法获取结构体有多少个字段，然后通过Field方法传递索引的方式，循环获取每一个字段，然后打印出他们的名字。
同样的对于方法也类似，这里不再赘述。

### 修改字段的值

假如我们想在运行中动态的修改某个字段的值有什么办法呢？一种就是我们常规的有提供的方法或者导出的字段可以供我们修改，还有一种是使用反射，这里主要介绍反射。
~~~
func main() {
    x:=2
    v:=reflect.ValueOf(&x)
    v.Elem().SetInt(100)
    fmt.Println(x)
}
~~~
以上就是通过反射修改一个变量的例子。

因为reflect.ValueOf函数返回的是一份值的拷贝，所以前提是我们是传入要修改变量的地址。
其次需要我们调用Elem方法找到这个指针指向的值。
最后我们就可以使用SetInt方法修改值了。

以上有几个重点，才可以保证值可以被修改，Value为我们提供了CanSet方法可以帮助我们判断是否可以修改该对象。

我们现在可以更新变量的值了，那么如何修改结构体字段的值呢？大家自己试试。

### 动态调用方法

结构体的方法我们不光可以正常的调用，还可以使用反射进行调用。要想反射调用，我们先要获取到需要调用的方法，然后进行传参调用，如下示例：
~~~
func main() {
    u:=User{"张三",20}
    v:=reflect.ValueOf(u)

    mPrint:=v.MethodByName("Print")
    args:=[]reflect.Value{reflect.ValueOf("前缀")}
    fmt.Println(mPrint.Call(args))

}
type User struct{
    Name string
    Age int
}
func (u User) Print(prfix string){
    fmt.Printf("%s:Name is %s,Age is %d",prfix,u.Name,u.Age)
}
~~~
MethodByName方法可以让我们根据一个方法名获取一个方法对象，然后我们构建好该方法需要的参数，最后调用Call就达到了动态调用方法的目的。

获取到的方法我们可以使用IsValid 来判断是否可用（存在）。

这里的参数是一个Value类型的数组，所以需要的参数，我们必须要通过ValueOf函数进行转换。

关于反射基本的介绍到这里就结束了，下一篇再介绍一些高级用法，比如获取字段的tag，常用的比如把一个json字符串转为一个struct就用到了字段的tag。
