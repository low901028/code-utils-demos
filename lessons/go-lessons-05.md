切片也是一种数据结构，它和数组非常相似，因为他是围绕动态数组的概念设计的，可以按需自动改变大小，使用这种结构，可以更方便的管理和使用数据集合。

### 内部实现
切片是基于数组实现的，它的底层是数组，它自己本身非常小，可以理解为对底层数组的抽象。因为机遇数组实现，所以它的底层的内存是连续非配的，效率非常高，还可以通过索引获得数据，可以迭代以及垃圾回收优化的好处。

切片对象非常小，是因为它是只有3个字段的数据结构：一个是指向底层数组的指针，一个是切片的长度，一个是切片的容量。这3个字段，就是Go语言操作底层数组的元数据，有了它们，我们就可以任意的操作切片了。

### 声明和初始化

切片创建的方式有好几种，我们先看下最简洁的make方式。
~~~
slice:=make([]int,5)
~~~
使用内置的make函数时，需要传入一个参数，指定切片的长度，例子中我们使用的时5，这时候切片的容量也是5。当然我们也可以单独指定切片的容量。
~~~
slice:=make([]int,5,10)
~~~
这时，我们创建的切片长度时5，容量时10,需要注意的这个容量10其实对应的是切片底层数组的。

因为切片的底层是数组，所以创建切片时，如果不指定字面值的话，默认值就是数组的元素的零值。这里我们所以指定了容量是10，但是我们职能访问5个元素，因为切片的长度是5，剩下的5个元素，需要切片扩充后才可以访问。

容量必须>=长度，我们是不能创建长度大于容量的切片的。

还有一种创建切片的方式，是使用字面量，就是指定初始化的值。
~~~
slice:=[]int{1,2,3,4,5}
~~~
有没有发现，是创建数组非常像，只不过不用制定[]中的值，这时候切片的长度和容量是相等的，并且会根据我们指定的字面量推导出来。当然我们也可以像数组一样，只初始化某个索引的值：
~~~
slice:=[]int{4:1}
~~~
这是指定了第5个元素为1，其他元素都是默认值0。这时候切片的长度和容量也是一样的。这里再次强调一下切片和数组的微小差别。
~~~
//数组
array:=[5]int{4:1}

//切片
slice:=[]int{4:1}
~~~
切片还有nil切片和空切片，它们的长度和容量都是0，但是它们指向底层数组的指针不一样，nil切片意味着指向底层数组的指针为nil，而空切片对应的指针是个地址。
~~~
//nil切片
var nilSlice []int

//空切片
slice:=[]int{}
~~~
nil切片表示不存在的切片，而空切片表示一个空集合，它们各有用处。

切片另外一个用处比较多的创建是基于现有的数组或者切片创建。
~~~
slice := []int{1, 2, 3, 4, 5}
slice1 := slice[:]
slice2 := slice[0:]
slice3 := slice[:5]

fmt.Println(slice1)
fmt.Println(slice2)
fmt.Println(slice3)
~~~
基于现有的切片或者数组创建，使用[i:j]这样的操作符即可，她表示以i索引开始，到j索引结束,截取原数组或者切片，创建而成的新切片，新切片的值包含原切片的i索引，但是不包含j索引。对比Java的话，发现和String的subString方法很像。

i如果省略，默认是0；j如果省略默认是原数组或者切片的长度,所以例子中的三个新切片的值是一样的。这里注意的是i和j都不能超过原切片或者数组的索引。
~~~
slice := []int{1, 2, 3, 4, 5}
newSlice := slice[1:3]

newSlice[0] = 10fmt.Println(slice)
fmt.Println(newSlice)
~~~
这个例子证明了，新的切片和原切片共用的是一个底层数组，所以当修改的时候，底层数组的值就会被改变，所以原切片的值也改变了。当然对于基于数组的切片也一样的。

我们基于原数组或者切片创建一个新的切片后，那么新的切片的大小和容量是多少呢？这里有个公式：

对于底层数组容量是k的切片slice[i:j]来说
长度：j-i
容量:k-i
比如我们上面的例子slice[1:3],长度就是3-1=2，容量是5-1=4。不过代码中我们计算的时候不用这么麻烦，因为Go语言为我们提供了内置的len和cap函数来计算切片的长度和容量。
~~~
slice := []int{1, 2, 3, 4, 5}
newSlice := slice[1:3]

fmt.Printf("newSlice长度:%d,容量:%d",len(newSlice),cap(newSlice))
~~~
以上基于一个数组或者切片使用2个索引创建新切片的方法，此外还有一种3个索引的方法，第3个用来限定新切片的容量，其用法为slice[i:j:k]。
~~~
slice := []int{1, 2, 3, 4, 5}
newSlice := slice[1:2:3]
~~~
这样我们就创建了一个长度为2-1=1，容量为3-1=2的新切片,不过第三个索引，不能超过原切片的最大索引值5。

### 使用切片

使用切片，和使用数组一样，通过索引就可以获取切片对应元素的值，同样也可以修改对应元素的值。
~~~
slice := []int{1, 2, 3, 4, 5}
fmt.Println(slice[2]) //获取值
slice[2] = 10 //修改值
fmt.Println(slice[2]) //输出10
~~~
切片只能访问到其长度内的元素，访问超过长度外的元素，会导致运行时异常，与切片容量关联的元素只能用于切片增长。

我们前面讲了，切片算是一个动态数组，所以它可以按需增长，我们使用内置append函数即可。append函数可以为一个切片追加一个元素，至于如何增加、返回的是原切片还是一个新切片、长度和容量如何改变这些细节，append函数都会帮我们自动处理。
~~~
slice := []int{1, 2, 3, 4, 5}
newSlice := slice[1:3]

newSlice=append(newSlice,10)
fmt.Println(newSlice)
fmt.Println(slice)
//Output
[2 3 10]
[1 2 3 10 5]
~~~
例子中，通过append函数为新创建的切片newSlice,追加了一个元素10，我们发现打印的输出，原切片slice的第4个值也被改变了，变成了10。引起这种结果的原因是因为newSlice有可用的容量，不会创建新的切片来满足追加，所以直接在newSlice后追加了一个元素10，因为newSlice和slice切片共用一个底层数组，所以切片slice的对应的元素值也被改变了。

这里newSlice新追加的第3个元素，其实对应的是slice的第4个元素，所以这里的追加其实是把底层数组的第4个元素修改为10，然后把newSlice长度调整为3。

如果切片的底层数组，没有足够的容量时，就会新建一个底层数组，把原来数组的值复制到新底层数组里，再追加新值，这时候就不会影响原来的底层数组了。

所以一般我们在创建新切片的时候，最好要让新切片的长度和容量一样，这样我们在追加操作的时候就会生成新的底层数组，和原有数组分离，就不会因为共用底层数组而引起奇怪问题,因为共用数组的时候修改内容，会影响多个切片。

append函数会智能的增长底层数组的容量，目前的算法是：容量小于1000个时，总是成倍的增长，一旦容量超过1000个，增长因子设为1.25，也就是说每次会增加25%的容量。

内置的append也是一个可变参数的函数，所以我们可以同时追加好几个值。
~~~
newSlice=append(newSlice,10,20,30)
~~~
此外，我们还可以通过...操作符，把一个切片追加到另一个切片里。
~~~
slice := []int{1, 2, 3, 4, 5}
newSlice := slice[1:2:3]

newSlice=append(newSlice,slice...)
fmt.Println(newSlice)
fmt.Println(slice)
~~~
### 迭代切片
切片是一个集合，我们可以使用 for range 循环来迭代它，打印其中的每个元素以及对应的索引。
~~~
    slice := []int{1, 2, 3, 4, 5} 
    for i,v:=range slice{
        fmt.Printf("索引:%d,值:%d\n",i,v)
    }
~~~
如果我们不想要索引，可以使用_来忽略它，这是Go语言的用法，很多不需要的函数等返回值，都可以忽略。
~~~
    slice := []int{1, 2, 3, 4, 5}    
    for _,v:=range slice{
        fmt.Printf("值:%d\n",v)
    }
~~~
这里需要说明的是range返回的是切片元素的复制，而不是元素的引用。

除了for range循环外，我们也可以使用传统的for循环，配合内置的len函数进行迭代。
~~~
    slice := []int{1, 2, 3, 4, 5}    
    for i := 0; i < len(slice); i++ {
        fmt.Printf("值:%d\n", slice[i])
    }
~~~
### 在函数间传递切片
我们知道切片是3个字段构成的结构类型，所以在函数间以值的方式传递的时候，占用的内存非常小，成本很低。在传递复制切片的时候，其底层数组不会被复制，也不会受影响，复制只是复制的切片本身，不涉及底层数组。
~~~
func main() {
    slice := []int{1, 2, 3, 4, 5}
    fmt.Printf("%p\n", &slice)
    modify(slice)
    fmt.Println(slice)
}
func modify(slice []int) {
    fmt.Printf("%p\n", &slice)
    slice[1] = 10
}
~~~
打印的输出如下：
~~~
0xc420082060
0xc420082080
[1 10 3 4 5]
~~~
仔细看，这两个切片的地址不一样，所以可以确认切片在函数间传递是复制的。而我们修改一个索引的值后，发现原切片的值也被修改了，说明它们共用一个底层数组。

在函数间传递切片非常高效，而且不需要传递指针和处理复杂的语法，只需要复制切片，然后根据自己的业务修改，最后传递回一个新的切片副本即可，这也是为什么函数间传递参数，使用切片，而不是数组的原因。

关于多维切片就不介绍了，还有多维数组，一来它和普通的切片数组一样，只不过是多个一维组成的多维；二来我压根不推荐用多维切片和数组，可读性不好，结构不够清晰，容易出问题。
