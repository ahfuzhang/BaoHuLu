# QiWa RPC 模板：兼容迁移，尚未完成服务端验收

```sh
hulu -src=examples/DemoServer/proto/Demo.proto -template.dir=builtin:csharp/QiWa.rpc -out=build/generated/rpc
```

`service` 每个服务执行一次，`rpc` 每个 method 执行一次；输入均为统一 RenderContext。`partials/model.tpl` 在模板内组装旧属性，named template 重绑定根 `$`。没有 Go 语言适配器，也不扫描既有输出发现数据类型。

文件名为 `服务.Server.cs` 和 `服务.方法.Context.cs`。生成文件受 manifest 管理，人工修改会阻止覆盖，不应直接在受管理文件里开发业务。

**D16 审核阻塞：** 旧正文尚未改变，context 仍调用不存在的无参数 `Clone()`，并假设响应有 `Code`、`Message` 字段；业务尚未拆为人工 partial 回调，不能声称生成物已可编译。不同服务共享 method 名会明确报错，避免输出冲突符号。通用核心支持这一场景，但内置旧模板需要独立适配。

此包不提供 tests、benchmarks、project 映射；输入不能显式开启这些能力。文件级缺省 `off/off/on` 可直接使用；没有项目映射时隐式 `on` 不报错。数据类型需另选数据模板生成，并人工维护框架项目引用。

`Method.Client__rename.cs` 是保留的客户端草稿，不在配置中，不会输出。
