# Go 模板包

`hulu -src=input.proto -template.dir=builtin:golang -out=generated/go`。
外部目录与内置资源使用同一加载、规划、渲染、写入流程。

- `types/`：每个 enum/message 一份文件，message 同时声明写类型与 Readonly 类型。
- `methods/`：每个 message 按 From/To 编解码方向、Clone、Reset 拆文件。
- `tests/`：由文件头 `@gen_test_code=on` 开启。
- `benchmarks/`：由 `@gen_benchmark_code=on` 开启；builder 不依赖普通测试。
- `project/`：默认创建 go.mod（Go 1.23）与 .gitignore；存在则保留。
- `partials/`：语言映射、64 位布局参数、字段命名、literal、样本、helper 和文件头。

入口接收统一 RenderContext。`model` 是**模板函数**，在模板内构造旧正文所需的局部属性；
这些属性不是 Go 核心的语言模型。现有正文只增加外层包装，不重写算法。

`AllTypes.go` 放两个类型；`AllTypes.FromJSON.go`、`AllTypes.ToJSON.go` 等放方法。
YAML/URL 依赖闭包由中立分析提供；ToYAML 定义同包的 YAML 常量。
普通 Go 文件不依赖 jsonv2 实验；jsonv2 benchmark 使用构建条件，
compare 还要求 `-tags=with.compare`，时长由 `benchmark_seconds` 控制。

内置模板对保留名、字段/方法冲突、生成符号冲突、会被 Go 忽略的平台文件名报错。
旧模板不支持 repeated enum；现在写入前明确报错，修复正文需按 D16 单独审核。
YAML 私有注释常量仍沿用行号命名，多 proto 共包的常量消歧尚待 D16 审核。
普通测试保留旧版的 229 pass / 11 fail，不能把迁移等价说成全部历史 bug 已修复。

完整协议见 [TemplateBundles](../../doc/TemplateBundles.md)，验证记录见 [GenerationValidation](../../doc/GenerationValidation.md)。
