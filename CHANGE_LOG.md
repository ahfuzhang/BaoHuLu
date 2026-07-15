* 2026-07-15: v0.16.0
  - csharp:
    - bug fix: 当某个 message 中包含了不在当前 proto 文件中的 message 时，程序会生成一个空的 message 的处理代码。
    - url values 编码中：父 message 定义了 @UrlValues，但是子 message 未加上 @UrlValues，这时候这个属性应该是传染性的。

* 2026-07-12: v0.15.0
  - golang:
    - bug fix: FromJSON() 中，对 null 的处理有错
    - bug fix: FromYAML() 中，对行内的尾部注释的处理有错
  - csharp:
    - bug fix: FromJSON() 对 @AsArray 的两层嵌套的处理有错
    - bug fix: ToYAML() 对 @AsArray 的两层嵌套的处理有错

