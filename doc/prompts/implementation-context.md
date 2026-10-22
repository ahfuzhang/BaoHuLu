# 实施上下文（2026-10-12）

用户已授权按 `plan.md` 实施整个项目，更新文档、清理旧实现，并使用 AllTypes.proto 验证。**数据生成主线已迁移，整份计划尚未完成。** 不得把尚待 D16 review 的 RPC 正文改动标为完成。

## 基线与约束

- 工作目录：`/Users/ahfu/code/github.com/ahfuzhang/BaoHuLu`；开始时 HEAD `5ddaa30`、工作区干净；未提交或创建分支。
- 固定旧工具、旧/新产物和原始日志：`/tmp/baohulu-migration.x42EWD/`。临时目录不是长期依赖，重建基线应使用固定提交的独立工作目录，不回退当前工作区。
- 当前验证环境：macOS arm64 / Apple M5、Go 1.27.0、.NET SDK 10.0.101；项目目标仍为 go 1.23 / net10.0。
- `question.md` 已有用户回答：工具链、项目元数据沿用上一版；空类型不生成代码。保留回答。
- 统一 CLI：`hulu -src=... -template.dir=... -out=...`，内部 check 后生成。没有操作子命令、语言参数、dry-run、类型 manifest。
- proto 文件头三个开关默认 off/off/on。目标语言逻辑放在模板内，Schema/codegen 不允许语言专属后端。
- D16：已有 .tpl 实现正文只能机械拆分，外层包装可适配上下文；必要正文改变须单独审核。已提出 RPC 例外问题，尚未收到批准，不得自行改写。
- 文件编辑使用 apply_patch；不覆盖人工代码或旧 build/ 内容。新命令产物使用 `build/generated/`。

## 已实现

1. `internal/schema`、`internal/protofile/parse.go`、`internal/protoextensions/file.go`：中立模型、两遍引用解析、语义校验、注解、值依赖/全依赖、协议事实、Service/RPC 关系及稳定引用列表。空类型跳过，引用空类型明确报错。
2. `internal/codegen`：严格 YAML、内置/外部统一 FS、统一 RenderContext、命名模板、模板实现的 typed fn/* helper、通用布局算法、稳定输出计划与完整渲染、路径/名称冲突检查。
3. 受管理 writer：单元身份、source/template/version 元数据、文件哈希和符号、create-only 项目文件、人工改动保护、同单元 prune、目录锁、事务日志、回滚/启动恢复。中断后旧锁须人工检查，恢复拒绝覆盖并发编辑。
4. `templates/golang` 与 `templates/csharp`：类型、codec、Clone/Reset、tests、benchmark、project 全部迁入；旧 Go helper 首次模板化进入 partials，不存在新 Go 语言适配器。
5. 统一 CLI、帮助/迁移诊断、Makefile、README、Extensions、CHANGE_LOG、TemplateBundles/Migration/GenerationValidation 文档已更新。
6. RPC 以模板内兼容模型接入统一 service/rpc 上下文；Demo 能生成三份文件，客户端草稿不输出。**不宣称 RPC 编译或通用 handler 完成。**
7. 已删除旧 `internal/golang`、`internal/csharp`、重复 protofile/protoextensions/protocheck 实现、旧 embed 与配置草案。运行库公开路径不变。

## 已验证

- `go test ./cmd/... ./internal/... ./dependencies/... ./templates/...`、go vet、核心/CLI/parser race、git diff --check 通过。
- `make generated-build`：真实 CLI 生成 AllTypes 后，Go 模块与 C# 主/测试/benchmark/GrpcGen 编译通过；Go jsonv2 编译通过。
- `TestCapabilityMatrix`：两种语言各 8 种开关组合，项目 on 的全部组合真实编译通过；仅测试、仅 benchmark 均覆盖。
- 内置与外部同配置输出逐字节一致。
- 在仓库外执行内置 Go/C#/RPC 生成通过；Go 1.23.12 仓库测试及 AllTypes 普通生成模块编译通过。
- AllTypes Go 生成测试 229 pass / 11 fail；C# 158 pass / 3 fail，失败集合与旧版完全一致（历史 JSON null / YAML 问题）。未删除或跳过旧测试以制造全绿。
- Go 所有非测试声明/函数格式化 AST 一致；C# 372 项声明/方法语法 token 一致。
- C# AllTypes 旧/新 JSON+Protobuf 双向互验、Google.Protobuf oracle、Clone/Reset 通过。
- 新增 go-contract 独立模块，Go/C# 共享 AllTypes 向量双向 JSON/Protobuf、官方 oracle、Clone/Reset 通过。C# 高区间 uint64 map key/fixed64 JSON 有符号转换与 uint32/fixed32 解析缺陷，已验证新旧一致并在 runner 中显式报告，不声称这些边界互通成功。
- 增加模板内目标名称/别名/enum 文件名校验，启动事务恢复和 metadata 确定性测试；File.OptionValues 提供中立原始字节区间，benchmark proto 副本仅修改 namespace option，不误改同值字段。
- 计划已按实际完成能力勾选，更新目录及链接；冲突的 plan_v2.md 已删除。question.md 末尾 R-01–R-04 记录 D16、历史 codec 与生成成本的待审核事项。
- `testdata/template-blocks.json` 与 `TestPreservedTemplateBlocks` 固定 45 个旧正文区块的 SHA-256，Go/C#/RPC 均一致。
- Shapes fixture 的 Go/C# 真实编译通过。Go repeated enum 是旧 codec 缺陷，单列 fixture 验证生成前报错，不伪装成支持。
- Go 新旧交替各 5 次、1s benchmark：AllTypes/ValueTypes 八个 JSON/Protobuf 场景中位数变化均在 ±5% 内，allocs/op 不变。C# benchmark 仅 build，未运行性能统计。
- 生成成本单列：初次新 Go/C# 中位数 1222/950 ms；增加语言无关、每次 Render 独立的 memo 后为 590.16/491.76 ms，旧版复测 61.52/83.26 ms。仍显著增加，未验收为可接受。更多文件/事务持久化和模板计算需后续 profiling，详情及原始样本见报告。
- 详细失败名、性能表、复现命令见 `doc/GenerationValidation.md`。

## 复现入口

```sh
make test
make generated-build
BAOHULU_BUILD_GENERATED=1 go test ./internal/integration -run 'TestCapabilityMatrix|TestBuiltinExternalEquivalence' -v
BAOHULU_BASELINE_GO=/tmp/baohulu-migration.x42EWD/go-old-fixture BAOHULU_MIGRATED_GO=/tmp/baohulu-migration.x42EWD/go-new go test ./internal/integration -run TestGoBaselineDeclarations -v
dotnet run --project internal/integration/csharp-contract/Contract.csproj -p:GeneratedDir=/tmp/baohulu-migration.x42EWD/csharp-new -p:BaselineDir=/tmp/baohulu-migration.x42EWD/csharp-old -- /tmp/baohulu-migration.x42EWD/csharp-old-fixture /tmp/baohulu-migration.x42EWD/csharp-new
```

单独保存生成目录：`BAOHULU_GENERATED_TEST_DIR=/tmp/独立目录 go test ./internal/integration -run '^TestAllTypesGo$' -v`；C# 使用 TestAllTypesCSharp，两个语言目录不能共用。手工 replace/tidy 修改的 go.mod 是受保护 create-only，不是回归。

旧 benchmark 项目为 `csharp-old/Benchmarks/Demo.bench.csproj`；新项目为 `csharp-new/Benchmarks/AllTypes.Benchmarks.csproj`。

## 尚未完成与续接顺序

1. 最终核心/最低 Go/race/vet、make generated-build、16 组合、Go AST、C# 372 项及 45 区块哈希均已在 memo 优化后重跑通过。AllTypes 全量生成测试仍为基线失败集合。跨语言互通、外部目录执行、确定性/恢复基本测试已完成；可继续未齐的边界矩阵，不把历史计划所有方框一律勾选。
2. RPC 需单独批准正文变化：不存在的 req.Clone() 改 Clone(ref dst)，移除 rsp.Code/Message 假设，人工 partial handler，跨 service 同名方法符号消歧，然后真实 QiWa 编译。
3. Go repeated enum 正式 codec 支持、跨 proto YAML 按源行号命名的私有常量冲突，需要 D16 例外；当前明确限制，不修改受保护正文。
4. 生成耗时/源码体积已测，生成端仍需 profiling/优化或接受差异。完整 C# 性能、编译耗时与程序集体积、linux/amd64 和 AOT 矩阵尚未完成；Go 最低工具链 1.23.12 已验证。未运行项必须在报告中注明。
5. 最终更新本上下文和验证报告，再给出诚实交付摘要；不能宣称整份计划验收完成或承诺额度恢复后自动运行。
