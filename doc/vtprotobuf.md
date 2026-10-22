# Go Protobuf VT 编码路径

当前内置 Go 包在 `templates/golang/methods/ToProtobuf.go.tpl` 中保留既有 VT 编码实现，与 `ProtobufSize`、`ToProtobufByAppend`、`ToProtobuf` 一同生成到每个 message 的 `类型名.ToProtobuf.go`。解码入口为 `FromProtobuf` / `FromProtobufWithArena`，位于 `类型名.FromProtobuf.go`。

```sh
hulu -src=input.proto -template.dir=builtin:golang -out=新目录
```

不提供 `-go_out.with.vtprotobuf`，也不生成独立的 `.vtprotobuf.go`；不要依赖旧草案中的 `ProtobufSizeVT` / `FromProtobufVT` 名称。测试与 benchmark 由 proto 文件级 `@gen_test_code` / `@gen_benchmark_code` 控制。

本次仅迁移模板文件边界，不调整既有算法。公开方法与旧版 AST 对照结果见 [验证报告](GenerationValidation.md)，模板包规范见 [TemplateBundles](TemplateBundles.md)。

