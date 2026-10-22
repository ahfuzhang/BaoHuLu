# 统一生成器迁移与恢复

## 命令与开关

`hulu -src=<file> -template.dir=<bundle> -out=<dir>` 是唯一生成形态；帮助和版本查询除外，每次调用先校验，再生成。

| 旧入口 | 新入口 |
| --- | --- |
| generate / tu | 删除子命令 |
| check / xi | 无独立校验替代入口；旧调用报错，不自动改成生成 |
| -go_out / -csharp_out | -template.dir=builtin:golang / builtin:csharp，加 -out |
| -src.csharp_template.dir / -dst.csharp_template.out_dir | -template.dir / -out |
| -with.test / -go_out.with.test / -csharp_out.with.test | proto 头部 @gen_test_code=on/off，默认 off |
| benchmark 对应开关 | @gen_benchmark_code=on/off，默认 off |
| -with.project / -no.csproj.file | @gen_project_file=on/off，默认 on；关闭整个 project，不仅主 csproj |

没有 dry-run、types.manifest、language 或操作子命令。生成到临时目录审阅差异；一个 proto 输出两种语言需调用两次。

## 首次迁移

1. 保留旧产物备份，将新代码生成到全新目录（Makefile 使用 build/generated）。
2. 编译独立模块/项目、确认业务引用后，再替换旧目录；不要把单文件旧版和拆分新版同时加入编译。
3. 转移人工项目配置、业务逻辑和依赖。生成器不会根据文件头猜测旧文件归属，也不删除清单外文件。
4. RPC 兼容包尚未通过业务拆分的 D16 审核，参见该包 README；不要将它当作已完成的通用 handler。

| 旧产物示例 | 新产物布局 |
| --- | --- |
| Go `demo.go` | `AllTypes.go`、`ReadonlyAllTypes` 同文件；其他 message/enum 各自独立，方法在 `AllTypes.FromJSON.go` 等文件 |
| Go `demo.AllTypes.yaml.go` | `AllTypes.FromYAML.go`、`AllTypes.ToYAML.go` |
| Go `demo_test.go` / `demo_timing_test.go` | 各 message 的独立 test / benchmark 文件，样本 builder 单独生成 |
| C# `Demo.AllTypes.cs` / `Demo.ReadonlyAllTypes.cs` | `AllTypes.cs` 两种类型声明及 `AllTypes.FromJSON.cs` 等 partial 方法文件 |
| C# `Demo.Enums.cs` | 每个 enum 独立文件，例如 `Status.cs` |
| C# `Demo.*.yaml.cs` / `*.UrlValues.cs` | 各 message 的 From/To 两个方向文件 |

上表仅用于人工迁移，不授权工具删除旧目录内的文件。不要只替换一个旧文件后保留其余旧 codec；应在备份后切换整套产物。

旧 `internal/golang`、`internal/csharp`、`internal/protocheck`、旧语言模型和重复注解收集器已删除；运行库 `dependencies/golang` 保留。两份文档配置草案及互相冲突的旧 `plan_v2.md` 删除，实际包配置是唯一来源。原有 build/ 和人工示例文件没有被批量删除。

## 文件归属

每个输出根的 `.hulu-manifest.json` v1 保存多个单元：

- 单元 ID：对“源文件相对于输出根的路径 + 内置包名或外部包相对路径”进行哈希。整棵工作目录迁移不改变相对关系；单独移动输入/外部包被视为新单元，并保守报所有权冲突。
- metadata：生成器版本、源内容哈希、配置/已加载模板内容哈希；没有运行时间戳。
- 文件：相对路径、内容哈希、模板、代码部分、符号 ID、create-only 标记。不包含跨任务语言类型导出。

类型、方法、测试、benchmark、RPC 是 managed；存在且属于当前单元且未手改时才覆盖。配置变化、删除类型、关闭能力、模板返回空时，只删除该单元原先拥有且未手改的 managed 文件。

project 下文件（包括 go.mod、csproj、.gitignore）是 create-only：存在则保留，关闭能力也保留并提示。启用新功能后需要人工更新已存在项目的依赖。不同单元不能争用已被归属的共享文件；共用目录时应指定一个项目所有者，其余输入设 project=off。

无归属文件不能覆盖，即使内容与新输出一致；清单外的项目文件只保留而不自动认领。不提供 force 开关。手改 managed 文件后应把改动迁入模板/独立人工文件，而不是删除清单强制生成。

## 失败与恢复

完整渲染和所有权预检在提交前完成。输出根以 `.hulu-lock` 互斥，普通错误会释放锁；提交有 `.hulu-transaction.json` 日志及原内容快照，逐文件原子替换，清单最后提交。多目录不具有原子性。

进程崩溃后：先备份整个输出根，确认锁中 PID 对应的生成任务已停止；人工删除遗留锁，再运行原命令。若存在日志，工具先恢复旧版本，然后返回“已恢复，请检查并重跑”，不会在同次运行继续生成。若文件与原版本/计划版本均不一致，恢复拒绝覆盖人工并发修改，保留日志供人工处理。

没有单独恢复 CLI 参数；核心 `codegen.Recover` 也可由仓库内维护工具调用。日志含源文件内容，按 0600 写入，清理前注意备份权限。磁盘故障、非协作并发改动仍可能需要人工恢复；不宣称整个目录在崩溃瞬间原子切换。

## 有意收紧和遗留限制

前向 enum/message 统一解析。oneof/optional/required/group/extend/extensions/stream、非法字段编号/map key、重复定义、重复路径等在生成前报错。import 仅警告，不实现完整 protobuf 文件图。

空类型不生成；引用空类型报错。语言模板额外检测关键字/冲突/特殊文件名。Go repeated enum 原本生成不可编译代码，现在明确拒绝，未借迁移改写 codec。多输入 YAML 常量消歧、RPC 通用业务回调和剩余历史 codec 问题需要独立 review。

Go package/module、C# namespace/项目符号、enum 文件名及别名中的不安全源码字符也在写入前检查。JSON/YAML 别名不被限制为语言标识符（连字符、点、Unicode 可保留），但旧模板无法正确转义的引号、反斜杠、控制字符等明确报错，不输出损坏源码。
