
# particular case 特例

本文档讲述某些与其他 proto 代码生成工具不同的部分。

## bool 类型的 map key

JSON 标准中，并不允许 bool 类型为 key.
但是 protobuf 是允许的： `map<bool, string>`

* 为此：当 proto 文件中出现 bool 类型的 key 时：
  - 序列化：变成字符串为 key 的 map，key 的名字分别是 "true" 和 "false"
  - 反序列化：如果 key 是 "true"，则反序列化为 mapName[true] = xx，否则放到 mapName[false] = xx

## bytes 类型
当 proto 文件中出现 bytes 类型，json 中并没有对应的类型。

* 因此：
  - 序列化：会把 bytes 类型序列化为 base64 string
  - 反序列化：把 base64 string 解码为 byte[] 数组

## 浮点数特殊处理

当一个浮点数正好等于取整后的值时，按照整形处理：

例如：
`float64(5.0)==float64(int64(5))`

