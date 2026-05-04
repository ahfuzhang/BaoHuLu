
to dear AI:

# 目标

复现项目 https://github.com/planetscale/vtprotobuf 的 golang 代码生成功能，以便实现高性能的 protobuf 中对 message 的序列化和反序列化。

# 背景信息

* vtprotobuf 项目源码：
  - ./build/vtprotobuf/
* 使用 bufbuild 生成的 vtprotobuf 的代码:
  - build/buf/golang/vtprotobuf/Demo_vtproto.pb.go
    - SizeVT(): 计算数据序列化后的大小, 代码 797 ~ 993 行
    - MarshalToVT(): 把类型序列化到一个数组，代码 126 ~ 547 行
    - UnmarshalVT(): 把二进制数据反序列化到类型，代码 1286 ~ 3137 行

# 输出

## ./internal/golang/vtprotobuf.go.tpl

生成一个模板文件，为类型生成如下方法：
* ProtobufSizeVT(): 计算类型序列化为 protobuf 二进制格式后的大小。等同于 vtprotobuf 库中的 SizeVT()
* ToProtobufVT(): 把类型序列化为 protobuf 二进制格。等同于 vtprotobuf 库中的 MarshalToVT()
* FromProtobufVT(): 把二进制反序列化到类型。等同于 vtprotobuf 库中的 UnmarshalVT()

## ./cmd/hulu/main.go

增加命令行参数 `-go_out.with.vtprotobuf`，当存在这个选项时，使用模板文件 ./internal/golang/vtprotobuf.go.tpl 来输出每个 message 的额外的 `ProtobufSizeVT()`, `ToProtobufVT()` 和 `FromProtobufVT()` 三个方法。这三个方法输出到文件：${proto_file_name}.vtprotobuf.go

# 处理过程

* 读取例子文件 build/buf/golang/vtprotobuf/Demo_vtproto.pb.go
* 读取例子类型： AllTypes
* 读取例子类型的如下三个方法：
  - SizeVT(): 计算数据序列化后的大小, 代码 797 ~ 993 行
  - MarshalToVT(): 把类型序列化到一个数组，代码 126 ~ 547 行
  - UnmarshalVT(): 把二进制数据反序列化到类型，代码 1286 ~ 3137 行
* 观察 vtprotobuf 库如何处理 19 种 protobuf 的数据类型。
* 生成模板文件 ./internal/golang/vtprotobuf.go.tpl
  - 注意：每个 message 对应的 struct 已经在 ${proto_file_name}.go 中生成好了。
  - 只需要额外增加文件 ${proto_file_name}.vtprotobuf.go
  - 新增的文件中，新增了三个方法：
    - Readonly${MessageName} 类型：增加 `FromProtobufVT()` 方法
    - {MessageName} 类型：增加 `ProtobufSizeVT()`, `ToProtobufVT()` 方法

## 其他约束
* 对于 string 类型的反序列化，使用 unsafe 来直接引用 buffer 中的内容，避免拷贝


