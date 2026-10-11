# Go 模板拆分预览

按 [Go 配置草案](../../doc/prompts/golang_tpl_config.yaml) 拆分现有模板，供审核文件职责与阅读体验。
[tpl_config.yaml](tpl_config.yaml) 中的 19 个映射均有实际模板，公共文件头放在 `partials/`。

本目录尚未接入 CLI。当前 `hulu tu -go_out` 仍使用 `internal/golang/templates/`；这里是一版基于现有数据和 helper 的模板预览，不代表统一生成引擎已经完成。

## 目录与职责

```text
templates/golang/
├── tpl_config.yaml
├── types/
│   ├── message.go.tpl           # 字段常量、写类型、Readonly 类型
│   └── enum.go.tpl              # 枚举声明及枚举值
├── methods/
│   ├── FromJSON.go.tpl          # FromJSON、fromJSONValue、fromJSONArray
│   ├── ToJSON.go.tpl            # 对象、AsMap、AsArray 编码
│   ├── FromProtobuf.go.tpl      # FromProtobuf、FromProtobufWithArena
│   ├── ToProtobuf.go.tpl        # 大小计算、append / VT 编码、公共入口
│   ├── FromYAML.go.tpl
│   ├── ToYAML.go.tpl            # 同时定义两方向共用的 YAML key 常量
│   ├── FromURLValues.go.tpl
│   ├── ToURLValues.go.tpl
│   ├── Reset.go.tpl             # 写类型与 Readonly 类型的 Reset
│   └── Clone.go.tpl
├── tests/
│   └── message_test.go.tpl      # 单个 message 的样本构造与测试
├── benchmarks/
│   ├── builder.go.tpl           # benchmark / compare 共用的样本构造
│   ├── message_bench_test.go.tpl
│   ├── jsonv2_bench_test.go.tpl  # goexperiment.jsonv2 构建条件
│   └── compare_test.go.tpl      # with.compare && goexperiment.jsonv2
├── project/
│   ├── go.mod.tpl
│   └── gitignore.tpl
└── partials/
    └── header.go.tpl            # 生成标记及 package 声明
```

每个模板直接包含相应的方法体，没有再把字段分支拆成大量子模板。`FromJSON`、`ToProtobuf` 仍各约 980 行，便于先评估按编解码方向聚合是否足够清晰。

## 生成文件的样子

以 `AllTypes` 为例，类型和方法文件是 `AllTypes.go`、`AllTypes.FromJSON.go`、`AllTypes.ToJSON.go`、`AllTypes.FromProtobuf.go`、`AllTypes.ToProtobuf.go`、`AllTypes.Reset.go`、`AllTypes.Clone.go`。YAML / URL Values 文件按注解及现有引用传播规则生成。

所有生成的 Go 文件放在同一个 package 目录；`types/`、`methods/` 等只是模板分类目录。测试和 benchmark 的输出名称取配置中的映射，例如 `AllTypes_test.go`、`AllTypes_bench_builder_test.go`。

每个文件有自己的 import。字段相关的 import 保留占位引用，避免空 message 或不同字段组合导致未使用导入。类型、字段常量只定义一次；YAML key 和注释常量由 `ToYAML` 文件定义，`FromYAML` 通过同包引用使用。

## 预览版输入约定

所有入口使用同一种根上下文，按当前渲染对象填充字段：

| 根字段 | 用途 |
| --- | --- |
| `Package` | 生成代码的 Go package 名称 |
| `Msg` | 当前 message，沿用现有 `golang.MsgTpl`，包含字段布局及注解 |
| `Enum` | 当前 enum，沿用现有 `golang.EnumTpl` |
| `Project` | `ModulePath`、`Version`、`WithTest`、`WithBench`、`WithDecimal` |

模板中的 `with .Msg` / `with .Enum` 只选择当前对象，不再遍历整个 proto 的全部类型。读取单个模板即可看到这个类型对应的完整代码部分。

渲染时需先注册 `internal/golang/golang.go` 中现有代码、测试、benchmark、YAML 和 URL Values 的 template helpers，再加载 `partials/` 和配置引用的模板。模板名称采用包内相对路径，例如 `methods/FromJSON.go.tpl`。本轮保留已有 helper 调用；语言无关的 `RenderContext` 以及 helper 迁入公共子模板仍属于后续改造。

`project/go.mod.tpl` 已从 `%s` 格式串改为 `text/template`，保留现有依赖版本与开发版本回退规则；`project/gitignore.tpl` 是新增的静态模板。

## 本地预览

本次校验生成的 Demo 文件位于 [build/template-preview/golang/demo](../../build/template-preview/golang/demo/)，可以先看 [AllTypes.go](../../build/template-preview/golang/demo/AllTypes.go)、[AllTypes.Reset.go](../../build/template-preview/golang/demo/AllTypes.Reset.go) 和 [AllTypes.ToProtobuf.go](../../build/template-preview/golang/demo/AllTypes.ToProtobuf.go)。这些是本地忽略目录下的预览产物，不是新的 CLI 入口。

校验使用临时适配代码复用现有生成器的数据和 helper，覆盖 Demo、空 message、单标量、decimal、嵌套 message 和 enum-only，并分别生成 tests-only / bench-only 组合。四组产物共 823 个类型、常量及函数声明与旧模板逐项一致；普通代码及启用 jsonv2 / compare 的代码均编译通过。仓库已有测试通过。

生成的测试尚有原有失败：Demo 为 229 项通过、11 项 JSON `null` 检查失败；边界样例为 81 项通过、5 项 `null` / decimal 检查失败。同样输入用旧模板重新生成后，通过数与失败项完全一致，本轮保留这些行为。
