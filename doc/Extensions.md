
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

