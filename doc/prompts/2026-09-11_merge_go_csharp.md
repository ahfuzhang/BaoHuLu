# 背景

当前项目是一个 proto 文件的代码生成工具，你可以理解为是一个更好的 protoc 或者 bufbuild/buf.
一开始，工具只专注于生成高性能的 golang/csharp 的序列化代码。
开发中发现，golang 的代码生成和 csharp 的代码生成，大多数情况下都是一样的。
我希望:
* 各种编程语言提供各自的 *.tpl 文件，使用 golang template 来语法来定义代码生成的逻辑。
   - 例外：为了兼容以前的命令行参数 `-go_out` , `-csharp_out`，我可能会使用 `// go:embed` 的模式来嵌入 tpl 文件到二进制中。但是 golang 代码本身进行不处理 golang/csharp 两种代码生成的逻辑。
* 生成文件的行为统一
  - csharp 对一个 proto 文件生成多个对应的 .cs 文件
  - golang 中，把多个 message 放到一个 .go 文件中，没有按照 message 的定义来划分。
  - 后续，我希望都修改为按照 message 切分文件。虽然文件的数量多了，但是每个文件的内容都很单一和清晰。

# 相关信息

* 例子 proto 文件
  - `examples/DemoServer/proto/Demo.proto`
  - 这个文件就是穷尽所有的类型，以便测试对每种类型处理的逻辑是否正常。
  - `AllTypes` 为了验证所有可能在 proto 文件中定义出来的类型
  - `ValueTypes` 仅仅只对值类型进行测试，这是大多数 proto 文件的常态

## golang 代码生成相关

* 主要代码目录:
  - `internal/golang`
* 代码模板目录：
  - `internal/golang/templates`
* 生成的代码目录
  - `build/golang/DemoServer`

## csharp 代码生成相关

* 主要代码目录:
  - `internal/csharp`
* 代码模板目录：
  - `internal/csharp/templates`
* 生成的代码的目录
  - `build/csharp/QiWa.rpc`

# 改动细节

* 根目录建立目录 `./templates`
  - 建立子目录 `golang`， 把原来 `internal/golang/templates` 目录的模板挪过去
  - 建立子目录 `csharp`，把原来 `internal/csharp/templates` 目录的模板挪过去
  - 可能需要在根目录引入一个 xx.go 来做 embed，否则子目录无法访问到 templates 目录。
* 原来分别处理 golang 和 csharp 的代码，合并成语言无关的
  - 对 proto 文件解析的各种结构体的定义，统一为一套
  - 把对语言要求生成的逻辑，转移到 .tpl 文件中处理
  - (尽量避免这种情况) 如果一定要在 golang 代码中生成某种语言的字符串，然后再把字符串传递到 tpl 文件中
    - 把这些部分定义为字符串的各种 util 函数，然后 tpl 文件中调用这个工具函数

# 最终效果

* 只需要指定 proto 文件，指定语言的模板目录，就能生成对应语言的数据序列化代码
  - 而不是像现在一样， golang 和 csharp 各有一套代码负责前期处理
* 所有的 message 的定义，都是独立的一个文件
  - 例如 AllTypes 类型会生成  AllTypes.go 或者 AllTypes.cs， 里面包含了 AllTypes 和 ReadonlyAllTypes 的 struct 定义。
* 每一个方法，生成一个文件
  - 例如 AllTypes 会生成如下文件
    - AllTypes.FromProtobuf.go
    - AllTypes.ToProtobuf.go
    - AllTypes.FromJSON.go
    - AllTypes.ToJSON.go
    - AllTypes.FromYAML.go
    - AllTypes.ToYAML.go
* template 文件本身也定义一个规范
  - 类型的模板
  - 各个方法的模板
  - 测试用例的模板
  - benchmark 的模板
* 有 rpc 的情况，做一个规划
  - 大多数情况下，我们只需要类型和序列化代码，不需要 rpc
  - rpc 代码的生成，通常需要针对 rpc 场景的额外 tpl 文件
  - rpc 代码生成后，需要比较容易引用各种 struct

