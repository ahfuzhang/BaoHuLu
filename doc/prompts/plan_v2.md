# 语言无关代码生成改造计划（精简版）

目标：只指定 proto 和模板目录即可生成代码；Go 代码不再包含 Go/C# 两套前处理，语言差异全部在 `.tpl` 中。

需求来源：[2026-09-11_merge_go_csharp.md](2026-09-11_merge_go_csharp.md)。详细论证见 [plan.md](plan.md)，本文只保留三件事：配置格式、要改的代码、模板如何拆分。

## 1. 配置文件格式

每个模板目录根部放一个 `tpl_config.yaml`，内容是“代码的哪一部分 → 哪个模板文件”。路径都相对于模板目录。完整样例见 [golang_tpl_config.yaml](golang_tpl_config.yaml) 和 [csharp_tpl_config.yaml](csharp_tpl_config.yaml)。

```yaml
message: types/message.go.tpl          # 每个 message 一次，含写类型和读类型
enum: types/enum.go.tpl                # 每个 enum 一次

json:
  FromJSON: methods/FromJSON.go.tpl
  ToJSON: methods/ToJSON.go.tpl
protobuf:
  FromProtobuf: methods/FromProtobuf.go.tpl
  ToProtobuf: methods/ToProtobuf.go.tpl
yaml:                                  # 是否生成由 proto 注解决定
  FromYAML: methods/FromYAML.go.tpl
  ToYAML: methods/ToYAML.go.tpl
url_values:                            # 同上
  FromURLValues: methods/FromURLValues.go.tpl
  ToURLValues: methods/ToURLValues.go.tpl
lifecycle:
  Reset: methods/Reset.go.tpl          # 写类型和读类型的 Reset 在同一文件
  Clone: methods/Clone.go.tpl

tests:                                 # 键 = 输出路径，值 = 模板
  "{message}_test.go": tests/message_test.go.tpl
benchmarks:
  "{message}_bench_test.go": benchmarks/message_bench_test.go.tpl
project:
  go.mod: project/go.mod.tpl
  .gitignore: project/gitignore.tpl
```

规则：

- **输出文件名**：`message`、`enum` 生成 `类型名.扩展名`；编解码和生命周期生成 `类型名.配置键.扩展名`，例如 `AllTypes.FromJSON.go`。扩展名取模板名去掉 `.tpl` 后的部分。
- **tests、benchmarks、project**：键直接写输出路径。`{message}` 按类型展开，`{file}` 是 proto 文件名（不含扩展名）。没有占位符的条目每次只生成一次。
- **按类型选模板**：enum 只用 `enum` 模板。message 用 `message` 加所有已配置且启用的方法、生命周期、测试、benchmark 模板，不用 `enum` 模板。
- **启用条件**：tests、benchmarks 由命令行开关控制。YAML、URL Values 由 proto 注解和引用传播规则控制。其余方法默认生成。
- **同类方法合并**：一个 From 或 To 模板包含该方向的全部变体和私有辅助方法，不再单独配置。例如 `ToProtobuf` 含 `ProtobufSize`、`ToProtobufVT`、`ToProtobufByAppend`，`FromJSON` 含 `fromJSONValue`、`fromJSONArray`。
- **公共子模板**：放在 `partials/`，自动加载，不出现在配置里。
- **校验**：用严格 YAML 解码。未知顶层项、重复键、空路径、找不到的模板文件都报错。渲染结果只有空白时不生成文件。
- **统一输入**：所有模板接收同一个 `RenderContext`，字段为 Schema、File、Message、Enum、Service、RPC、Analysis、Options。模板自行取用需要的字段。
- **RPC**：放在独立目录，例如 `templates/csharp/QiWa.rpc/tpl_config.yaml`，只有 `service:` 和 `rpc:` 两项。默认不生成，需显式指定该目录。

## 2. 需要修改的代码文件

### 2.1 新增

| 文件 | 内容 |
| --- | --- |
| `templates/embed.go` | 只做 `//go:embed`，导出内置 Go 和 C# 模板的 `fs.FS` |
| `internal/codegen/bundle.go` | 读取并校验 `tpl_config.yaml`，加载 `partials/` 和被引用的模板 |
| `internal/codegen/context.go` | 定义统一的 `RenderContext` |
| `internal/codegen/plan.go` | 按 enum 或 message 展开模板，计算输出路径，检查重名和越界 |
| `internal/codegen/render.go` | 执行模板，提供 `render` 等少量无语言语义的函数 |
| `internal/codegen/write.go` | 先全部渲染再写入，已有项目文件不覆盖，失败不留半成品 |

### 2.2 改造

| 文件 | 改动 |
| --- | --- |
| [cmd/hulu/main.go](../../cmd/hulu/main.go) | 去掉 Go/C# 分支和文件创建，只保留参数解析。旧参数 `-go_out`、`-csharp_out` 转成“内置模板目录 + 输出目录”。新增 `-template.dir`、`-out`。`modifyProtoNamespace` 改由模板输出 |
| [internal/protofile/protofile.go](../../internal/protofile/protofile.go) | 删除 `GoType`、`ScalarProtoToGo`、`ProtoTypeToGo`、`FieldGoSize` 等语言映射。保留原始类型引用，先收集全部符号再解析引用，避免前向 enum 被误判为 message。保留 wire type 和 tag 数值，布局排序改为按模板给的尺寸计算 |
| [internal/protoextensions/extensions.go](../../internal/protoextensions/extensions.go) | 合并带行号和不带行号的两份解析，保留原始注解和位置 |
| [internal/protocheck/protocheck.go](../../internal/protocheck/protocheck.go) | 改为复用同一套解析和校验，使 `check` 与 `generate` 一致 |
| [internal/golang/golang.go](../../internal/golang/golang.go) | 逻辑迁入 `templates/golang/partials/` 和模板，迁完删除 |
| [internal/golang/go_mod.go](../../internal/golang/go_mod.go) | 已由 `project/go.mod.tpl` 取代，迁完删除 |
| [internal/csharp/csharp.go](../../internal/csharp/csharp.go) | 逻辑迁入 `templates/csharp/partials/` 和模板，迁完删除 |
| `internal/csharp/csharp_*.go` | 三个项目文件辅助代码由 `project/*.tpl` 取代，迁完删除 |
| `internal/*/templates/` | 模板迁到根目录 `templates/`，旧目录迁完删除 |
| [templates/csharp/QiWa.rpc/](../../templates/csharp/QiWa.rpc/) | 增加 `tpl_config.yaml`，把 `Service`、`Method` 两类文件映射为 `service`、`rpc`。`Method.Client__rename.cs` 是草稿，不纳入 |

不动：`dependencies/golang/`、`examples/`。

## 3. 模板如何拆分

`templates/golang/` 已有拆分预览，本节用来对照并补齐 C#。

### 3.1 Go

| 旧模板 | 新模板 |
| --- | --- |
| `message.go.tpl`（3137 行） | `types/message.go.tpl`、`types/enum.go.tpl`、`methods/` 下的 `FromJSON`、`ToJSON`、`FromProtobuf`、`ToProtobuf`、`Reset`、`Clone` |
| `yaml.go.tpl` | `methods/FromYAML.go.tpl`、`methods/ToYAML.go.tpl`。两方向共用的 key 常量放 `ToYAML` |
| `url.values.go.tpl` | `methods/FromURLValues.go.tpl`、`methods/ToURLValues.go.tpl` |
| `test.go.tpl` | `tests/message_test.go.tpl` |
| `timing.test.go.tpl` | `benchmarks/builder.go.tpl`、`benchmarks/message_bench_test.go.tpl`、`benchmarks/jsonv2_bench_test.go.tpl` |
| `compare.test.go.tpl` | `benchmarks/compare_test.go.tpl` |
| `go.mod.tpl` | `project/go.mod.tpl`，另增 `project/gitignore.tpl` |
| 每个文件开头的 package 和 import | `partials/header.go.tpl`，每个输出文件自己声明 import |

要点：

- 所有输出放在同一个 package 目录，`types/`、`methods/` 只是模板分类。
- 类型和字段常量只定义一次。每个文件自己写 import，不依赖其他文件。
- `FromJSON`、`ToProtobuf` 各约 980 行，先保持整块，之后再看是否要抽公共子模板。
- Go 类型映射、tag 字节展开、literal 等 `golang.go` 里的函数，迁成 `partials/` 里的命名子模板。

### 3.2 C#（待新建，结构与 Go 对应）

| 旧模板 | 新模板 |
| --- | --- |
| `message.cs.tpl`（1691 行） | `types/message.cs.tpl`（写类型、读类型、Tags、Wrapper）、`types/enum.cs.tpl`、`methods/` 下的 `FromJSON`、`ToJSON`、`FromProtobuf`、`ToProtobuf`（含 `ProtobufSize`）、`Reset`、`Clone` |
| `message.yaml.cs.tpl` | `methods/FromYAML.cs.tpl`、`methods/ToYAML.cs.tpl` |
| `url.values.cs.tpl` | `methods/FromURLValues.cs.tpl`、`methods/ToURLValues.cs.tpl` |
| `test.cs.tpl` | `tests/message.Tests.cs.tpl`（每个 message）、`tests/TestSupport.cs.tpl`（只生成一次） |
| `benchmark.cs.tpl` | `benchmarks/message.Benchmarks.cs.tpl`（每个 message）、`benchmarks/BenchmarkSupport.cs.tpl`（只生成一次） |
| `CSharp.Benchmarks.Program.cs.tpl` | `benchmarks/Program.cs.tpl` |
| 无（由 `modifyProtoNamespace` 改写） | `benchmarks/GrpcGen.proto.tpl` |
| `CSharp.csproj.tpl`、`CSharp.Tests.csproj.tpl`、`CSharp.Benchmarks.csproj.tpl`、`GrpcGen.csproj.tpl` | `project/main.csproj.tpl`、`project/tests.csproj.tpl`、`project/benchmarks.csproj.tpl`、`project/GrpcGen.csproj.tpl`，另增 `project/gitignore.tpl` |

要点：

- 方法文件使用同 namespace 下的 `partial struct`，与类型文件在同一个 assembly。
- 不把 `ReadonlyX` 改成 C# 的 `readonly struct`，方法签名保持不变。
- 测试和 benchmark 的公共辅助类只生成一次，避免每个 message 重复定义。
- 主项目继续排除 `Tests/` 和 `Benchmarks/`，benchmark 项目继续排除 `GrpcGen/`。
- `.csproj` 是否生成沿用旧 `-no.csproj.file` 的范围，即只禁止主项目。

## 4. 实施顺序

1. 把现有模板原样搬到根目录 `templates/`，加 `embed.go`，旧入口改读新位置。生成结果必须与搬迁前完全一致。
2. 实现 `internal/codegen`，用最小 proto 和 Go 预览模板验证配置加载、模板选择、输出路径。
3. 用 Go 和 C# 各取一小段类型加方法做原型，确认类型映射、递归、字段布局能只靠模板完成。原型不通过则调整配置和上下文，再继续。
4. 按第 3 节完成 Go 全部模板，再完成 C# 全部模板，并改造 `protofile`。每个阶段都要编译并运行生成的测试。
5. 改 `main.go` 接入新引擎，保留旧参数。
6. 迁移 QiWa.rpc 的服务端模板。
7. 对比新旧产物的 API、协议和性能，通过后删除 `internal/golang`、`internal/csharp`。

两个已知前提：生成文件名从 `demo.go` 变为 `AllTypes.go` 等，旧目录需要重新生成到新目录再切换，不能原地叠加。生成的 Go 测试目前有一批 JSON `null` 检查在旧模板下就失败，迁移时保持不变，不作为回归。
