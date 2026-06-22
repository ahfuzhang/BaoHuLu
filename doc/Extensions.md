
# 扩展语法

此文档约定在 proto 文件中使用的扩展语法。

* 扩展语法写在注释中
  - 就算使用传统的 protoc 工具，也仍然兼容。
* 注释的格式如下:

```
// @KeyWord=Value
```

* 注释必须单独一行
* 预先定义多个 KeyWord
  - 每个 KeyWord 对应着自己特有的取值
  - KeyWord 不区分大小写
  - `@`符号必须紧贴 KeyWord
  - 忽略 `=` 前后的空白字符
  - 忽略 `Value` 前后的空白字符

# 细节

## message 扩展语法

### @Deprecated

message 上存在此扩展信息时，相当于没有这个 message，生成代码时跳过此 message.

### @UrlValues

当存在这个标签时，为类型提供 url encode 的序列化和反序列化支持
* ${message_name} 的类型上支持 ToURLValues() 方法
  - 逐个字段进行格式化
  - key 的名字与 json 的名字一样
  - value 部分格式化为字符串
    - 如果是 bytes 类型，进行 base64 encode
    - 如果是 数组类型，写成连续的多个 key 的格式： k1=value1&k1=value2
    - 如果是 map 类型，先把所有的 key-value 转换为 `k1=v1&k2=v2` 这样的字符串，然后整个字符串再 url encode 后作为 value 赋值给 key
* Readonly${message_name} 的类型上支持 FromURLValues() 方法
  - 对数据进行 url decode
  - 逐个字段赋值到 Readonly${message_name} 中的字段中
  - 对于数组类型，把 key 存在重复的数据作为数组处理
  - 对于 map 类型：对 value 部分进行 url decode，然后按照 key-value 的格式写入
  - 对子嵌入的 message 类型：把 value 部分作为子类型的 FromPostForm() 方法的输入。
  - 对于 bytes 类型，认为内部是 base64 编码的

### @yaml

当存在这个标签是，为类型提供 yaml 格式的序列化和反序列化。
* ${message_name} 的类型上支持 ToYAML() 方法
  - 函数原型为： `func (t *TypeName) ToYAML(dst []byte, indent int) []byte`
    - indent 参数为缩进的空格数
  - 缩进为两个空格，下一级缩进在上一级缩进的基础上增加两个空格
  - 逐个字段追加到 dst 中，不使用任何 yaml 的代码库
* Readonly${message_name} 的类型上支持 FromYAML() 方法
  - 函数原型为： `func (t *ReadonlyTypeName) FromYAML(src []byte) error`
  - 逐行读取，根据前导的空格来逐层解析
  - 不使用任何 yaml 的代码库

### @AsMap

当存在这个标签时，把整个 message 理解为 map.
主要为了解决：
* protobuf 不支持类似的定义： `repeated map<string, string> data = 2;`
* 如下的定义会因为 `@AsMap` 标签而重新理解

```protobuf
// @AsMap
message MapType {
  map<string, string> labels = 1;
}

message Series {
  repeated MapType data = 1;
}
```

上面的 proto 能够描述下面的 json 数据：

```json
{
  "data":[
    {"label1":"value1", "label2":"value2"}
  ]
}
```

struct 的结构仍然不变，只是对 json 的序列化和反序列化做特殊处理:

```go
type MapType struct {
  Labels map[string]string
}

type Series struct {
  Data []MapType
}
```

* 注意：只影响 FromJSON() 和 ToJSON() 的行为，其他不变

### @AsArray

当存在这个标签时，把整个 message 理解为数组.
主要为了解决：
* protobuf 不支持类似的定义： `map<string, []string> data = 2;`
* 如下的定义会因为 `@AsArray` 标签而重新理解

```protobuf
// @AsArray
message ArrayType {
  repeated string items = 1;
}

message OneToMany {
  map<string, ArrayType> one_to_many = 2;
}
```

上面的 proto 能够描述下面的 json 数据：

```json
{
  "one_to_many": {
    "key1":["item1", "item2"],
    "key2":["item3", "item4"]
  }
}
```

struct 的结构仍然不变，只是对 json 的序列化和反序列化做特殊处理:

```go
type ArrayType struct {
  Items []string
}

type OneToMany struct {
  OneToMany map[string]*ArrayType
}
```

* 注意：只影响 FromJSON() 和 ToJSON() 的行为，其他不变


## field 扩展语法

### @Deprecated

field 上存在此扩展信息时，相当于没有这个 field，生成代码时跳过此 field.

### @VarName

* 用于告诉代码生成工具，在生成类型的成员时，使用给定的变量名

例如:

```protobuf
message Child{
    int32 child_id = 1;
}
```

通常生成以下代码:

```go
type Child struct{
    ChildId int32
}
```

而 `Id` 这个词不符合 golang 的 lint 规范，应该使用 `ID`。
此时可以这样处理：

```protobuf
message Child{
    // @VarName=ChildID
    int32 child_id = 1;
}
```

代码生成工具发现 `// @VarName=ChildID` 这样的注释，会使用 VarName 后的值作为成员的名字。

### @jsonName

* 此扩展用于重新定义 json 序列化和反序列化时候对应的 key 的名字。
* 当存在 `// @jsonName=xxx` 时：
  - json 名字的常量定义中，使用注释中的名字。（注意检查名字的唯一性，名字不唯一时报错）
  - golang 对应的 struct 中，每个成员的注解内，使用扩展语法内的名字

### @yamlName

* 此扩展用于定义 yaml 序列化和反序列化时候对应的 key 的名字
* 某个 message 只要出现了 @yamlName 的扩展，就额外定义一组字符串常量 `NameOfYamlField${XX}`，便于引用。
* 对于 golang，每个 struct 的成员后面，加上 yaml 的注解信息：`json:"${XX},omitempty", yaml:"${XX}"`

### @tag

* 此扩展用于定义任意的 tag
* 语法为： `// @tag=${Name}:${value}`
* 对于 golang:
  - 每个 struct 后的成员，在 json 这个 tag 之后增加 `json:"${XX},omitempty", ${Name}:"${Value}"`

### @formName

* 此扩展用于定义使用 `application/x-www-form-urlencoded` 进行编码的表单的名字

### @decimal

* 此扩展，便于在金融领域使用 decimal 类型来表示金额
* 语法为 `// @decimal=round:5`
* 限制：必须用于 double 类型的字段上方
* 对于 golang:
  - 使用库 github.com/govalues/decimal 中的 decimal 类型。代码模版会把 double 替换为 decimal
* 对于 csharp
  - 使用 csharp 自带的 decimal 类型
* 精度处理：
  - `round:5` 这一段的含义是，最大的小数位数为 5
  - 序列化时：根据 round 的小数精度，在指定的位数上做四舍五入. 然后按照 double 类型来序列化
  - 反序列化时：先按照 double 类型来反序列化，然后按照 round 要求的精度来四舍五入，然后再赋值给 decimal 类型。

## method 扩展语法

### @path

在 rpc 上，可以通过 @path 来指定 api 的访问路径。

```protobuf
service Demo {
  // @path=/api/v1/login
  rpc Login(LoginRequest) returns (LoginResponse) {}
  rpc GetUserInfo(GetUserInfoRequest) returns (GetUserInfoResponse) {}
  rpc SetUserTags(SetUserTagsRequest) returns (SetUserTagsResponse) {}
}
```

当生成 service 的代码时，`@path` 中指定的路径变成分发的路径。

