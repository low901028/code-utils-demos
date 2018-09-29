# Go unsafe 包之内存布局
unsafe，顾名思义，是不安全的，Go定义这个包名也是这个意思，让我们尽可能的不要使用它，如果你使用它，看到了这个名字，也会想到尽可能的不要使用它，或者更小心的使用它。

虽然这个包不安全，但是它也有它的优势，那就是可以绕过Go的内存安全机制，直接对内存进行读写，所以有时候因为性能的需要，会冒一些风险使用该包，对内存进行操作。

### Sizeof函数
Sizeof函数可以返回一个类型所占用的内存大小，这个大小只有类型有关，和类型对应的变量存储的内容大小无关，比如bool型占用一个字节、int8也占用一个字节。
~~~
func main() {
    fmt.Println(unsafe.Sizeof(true))
    fmt.Println(unsafe.Sizeof(int8(0)))
    fmt.Println(unsafe.Sizeof(int16(10)))
    fmt.Println(unsafe.Sizeof(int32(10000000)))
    fmt.Println(unsafe.Sizeof(int64(10000000000000)))
    fmt.Println(unsafe.Sizeof(int(10000000000000000)))
}
~~~
对于整型来说，占用的字节数意味着这个类型存储数字范围的大小，比如int8占用一个字节，也就是8bit，
所以它可以存储的大小范围是-128~~127,也就是−2^(n-1)到2^(n-1)−1，n表示bit，int8表示8bit，int16表示16bit，其他以此类推。

对于和平台有关的int类型，这个要看平台是32位还是64位，会取最大的。比如我自己测试，以上输出，会发现int和int64的大小是一样的，因为我的是64位平台的电脑。
~~~
func Sizeof(x ArbitraryType) uintptr
~~~
以上是Sizeof的函数定义，它接收一个ArbitraryType类型的参数，返回一个uintptr类型的值。这里的ArbitraryType不用关心，他只是一个占位符，为了文档的考虑导出了该类型，但是一般不会使用它，我们只需要知道它表示任何类型，也就是我们这个函数可以接收任意类型的数据。
~~~
// ArbitraryType is here for the purposes of documentation only and is not actually
// part of the unsafe package. It represents the type of an arbitrary Go expression.
type ArbitraryType int
Alignof 函数
~~~
Alignof返回一个类型的对齐值，也可以叫做对齐系数或者对齐倍数。对齐值是一个和内存对齐有关的值，合理的内存对齐可以提高内存读写的性能，关于内存对齐的知识可以参考相关文档，这里不展开介绍。
~~~
func main() {
    var b bool
    var i8 int8
    var i16 int16
    var i64 int64

    var f32 float32

    var s string

    var m map[string]string

    var p *int32

    fmt.Println(unsafe.Alignof(b))
    fmt.Println(unsafe.Alignof(i8))
    fmt.Println(unsafe.Alignof(i16))
    fmt.Println(unsafe.Alignof(i64))
    fmt.Println(unsafe.Alignof(f32))
    fmt.Println(unsafe.Alignof(s))
    fmt.Println(unsafe.Alignof(m))
    fmt.Println(unsafe.Alignof(p))

}
~~~
从以上例子的输出，可以看到，对齐值一般是2^n,最大不会超过8（原因见下面的内存对齐规则）。Alignof的函数定义和Sizeof基本上一样。这里需要注意的是每个人的电脑运行的结果可能不一样，大同小异。
~~~
func Alignof(x ArbitraryType) uintptr
~~~
此外，获取对齐值还可以使用反射包的函数，也就是说：unsafe.Alignof(x)等价于reflect.TypeOf(x).Align()。

### Offsetof 函数

Offsetof函数只适用于struct结构体中的字段相对于结构体的内存位置偏移量。结构体的第一个字段的偏移量都是0.
~~~
func main() {
    var u1 user1

    fmt.Println(unsafe.Offsetof(u1.b))
    fmt.Println(unsafe.Offsetof(u1.i))
    fmt.Println(unsafe.Offsetof(u1.j))
}
type user1 struct {
    b byte
    i int32
    j int64
}
~~~
字段的偏移量，就是该字段在struct结构体内存布局中的起始位置(内存位置索引从0开始)。根据字段的偏移量，我们可以定位结构体的字段，进而可以读写该结构体的字段，哪怕他们是私有的，黑客的感觉有没有。偏移量的概念，我们会在下一小结详细介绍。

此外，unsafe.Offsetof(u1.i)等价于reflect.TypeOf(u1).Field(i).Offset

### 有意思的struct大小
我们定义一个struct，这个struct有3个字段，它们的类型有byte,int32以及int64,但是这三个字段的顺序我们可以任意排列，那么根据顺序的不同，一共有6种组合。
~~~
type user1 struct {
    b byte
    i int32
    j int64
}
type user2 struct {
    b byte
    j int64
    i int32
}
type user3 struct {
    i int32
    b byte
    j int64
}
type user4 struct {
    i int32
    j int64
    b byte
}
type user5 struct {
    j int64
    b byte
    i int32
}
type user6 struct {
    j int64
    i int32
    b byte
}
~~~
根据这6种组合，定义了6个struct，分别位user1，user2，…，user6，那么现在大家猜测一下，这6种类型的struct占用的内存是多少，就是unsafe.Sizeof()的值。

大家可能猜测1+4+8=13，因为byte的大小为1，int32大小为4，int64大小为8，而struct其实就是一个字段的组合，所以猜测struct大小为字段大小之和也很正常。

但是，但是，我可以明确的说，这是错误的。

为什么是错误的，因为有内存对齐存在，编译器使用了内存对齐，那么最后的大小结果就不一样了。现在我们正式验证下，这几种struct的值。
~~~
func main() {
    var u1 user1    
    var u2 user2    
    var u3 user3    
    var u4 user4    
    var u5 user5    
    var u6 user6

    fmt.Println("u1 size is ",unsafe.Sizeof(u1))
    fmt.Println("u2 size is ",unsafe.Sizeof(u2))
    fmt.Println("u3 size is ",unsafe.Sizeof(u3))
    fmt.Println("u4 size is ",unsafe.Sizeof(u4))
    fmt.Println("u5 size is ",unsafe.Sizeof(u5))
    fmt.Println("u6 size is ",unsafe.Sizeof(u6))
}
~~~
从以上输出可以看到，结果是：
~~~
u1 size is  16
u2 size is  24
u3 size is  16
u4 size is  24
u5 size is  16
u6 size is  16
~~~
结果出来了（我的电脑的结果，Mac64位，你的可能不一样），4个16字节，2个24字节，既不一样，又不相同，这说明：

### 内存对齐影响struct的大小
struct的字段顺序影响struct的大小
综合以上两点，我们可以得知，不同的字段顺序，最终决定struct的内存大小，所以有时候合理的字段顺序可以减少内存的开销。

内存对齐会影响struct的内存占用大小，现在我们就详细分析下，为什么字段定义的顺序不同会导致struct的内存占用不一样。

在分析之前，我们先看下内存对齐的规则：

> 对于具体类型来说，对齐值=min(编译器默认对齐值，类型大小Sizeof长度)。也就是在默认设置的对齐值和类型的内存占用大小之间，取最小值为该类型的对齐值。我的电脑默认是8，所以最大值不会超过8.
struct在每个字段都内存对齐之后，其本身也要进行对齐，对齐值=min(默认对齐值，字段最大类型长度)。这条也很好理解，struct的所有字段中，最大的那个类型的长度以及默认对齐值之间，取最小的那个。
以上这两条规则要好好理解，理解明白了才可以分析下面的struct结构体。在这里再次提醒，对齐值也叫对齐系数、对齐倍数，对齐模数。这就是说，每个字段在内存中的偏移量是对齐值的倍数即可。

我们知道byte，int32，int64的对齐值分别为1，4，8，占用内存大小也是1，4，8。那么对于第一个structuser1，它的字段顺序是byte、int32、int64，我们先使用第1条内存对齐规则进行内存对齐，其内存结构如下，内存布局中有竖线(|),用于每四个字节的分割，下同。

##### user1
bxxx|iiii|jjjj|jjjj
- user1类型，第1个字段byte，对齐值1，大小1，所以放在内存布局中的第1位。

- 第2个字段int32，对齐值4，大小4，所以它的内存偏移值必须是4的倍数，在当前的user1中，就不能从第2位开始了，必须从第5位开始，也就是偏移量为4。第2，3，4位由编译器进行填充，一般为值0，也称之为内存空洞。所以第5位到第8位为第2个字段i。

- 第3字段，对齐值为8，大小也是8。因为user1前两个字段已经排到了第8位，所以下一位的偏移量正好是8，是第3个字段对齐值的倍数，不用填充，可以直接排列第3个字段，也就是从第9位到第16位为第3个字段j。

现在第一条内存对齐规则后，内存长度已经为16个字节，我们开始使用内存的第2条规则进行对齐。根据第二条规则，默认对齐值8，字段中最大类型长度也是8，所以求出结构体的对齐值位8，我们目前的内存长度为16，是8的倍数，已经实现了对齐。

所以到此为止，结构体user1的内存占用大小为16字节。

##### user2
现在我们再分析一个user2类型，它的大小是24，只是调换了一下字段i和j的顺序，就多占用了8个字节，我们看看为什么？还是先使用我们的内存第1条规则分析。

bxxx|xxxx|jjjj|jjjj|iiii
按对齐值和其占用的大小，第1个字段b偏移量为0，占用1个字节，放在第1位。

第2个字段j，是int64，对齐值和大小都是8，所以要从偏移量8开始，也就是第9到16位为j，这也就意味着第2到8位被编译器填充。

目前整个内存布局已经偏移了16位，正好是第3个字段i的对齐值4的倍数，所以不用填充，可以直接排列，第17到20位为i。

现在所有字段对齐好了，整个内存大小为1+7+8+4=20个字节，我们开始使用内存对齐的第2条规则，也就是结构体的对齐，通过默认对齐值和最大的字段大小，求出结构体的对齐值为8。

现在我们的整个内存布局大小为20，不是8的倍数，所以我们需要进行内存填充，补足到8的倍数，最小的就是24，所以对齐后整个内存布局为

bxxx|xxxx|jjjj|jjjj|iiii|xxxx
所以这也是为什么我们最终获得的user2的大小为24的原因。
基于以上办法，我们可以得出其他几个struct的内存布局。

##### user3
iiii|bxxx|jjjj|jjjj
user4

iiii|xxxx|jjjj|jjjj|bxxx|xxxx
user5

jjjj|jjjj|bxxx|iiii
user6

jjjj|jjjj|iiii|bxxx
以上给出了答案，推到过程大家可以参考user1和user2试试。下一篇我们介绍通过unsafe.Pointer进行内存的运算，以及对内存的读写。
